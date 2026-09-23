package main

import (
	"context"
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

//go:embed static
var staticFiles embed.FS

type Target struct {
	Name string
	URL  string // whatever HTTP health checkpoint
}

type Status struct {
	Target    string    `json:"target"`
	Up        bool      `json:"up"`
	LatencyMs int64     `json:"latency_ms"`
	CheckedAt time.Time `json:"checked_at"`
}

type HostStatus struct {
	Connectivity     bool `json:"connectivity"`
	RebootRequired   bool `json:"reboot_required"`
	UpdatesAvailable int  `json:"updates_available"`
}

type StatusResponse struct {
	Services map[string]Status `json:"services"`
	Host     HostStatus        `json:"host"`
}

var (
	statuses   = make(map[string]Status)
	mu         sync.RWMutex
	hostStatus HostStatus
)

func checkOnce(t Target) Status {
	start := time.Now()

	response, err := http.Get(t.URL)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return Status{
			Target:    t.Name,
			Up:        false,
			LatencyMs: latency,
			CheckedAt: time.Now(),
		}
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return Status{
			Target:    t.Name,
			Up:        false,
			LatencyMs: latency,
			CheckedAt: time.Now(),
		}
	}

	up := response.StatusCode >= 200 && response.StatusCode < 300

	if !up {
		log.Printf(
			"response failed with status code %d and body: %s",
			response.StatusCode,
			body,
		)
	}

	return Status{
		Target:    t.Name,
		Up:        up,
		LatencyMs: latency,
		CheckedAt: time.Now(),
	}
}

func monitor(targets []Target, interval time.Duration, out chan<- Status, alerter Alerter) {
	for _, target := range targets {
		t := target
		go func() {
			var lastUp *bool
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				status := checkOnce(t)
				if lastUp != nil && *lastUp != status.Up {
					event := EventRecovery
					if !status.Up {
						event = EventDown
					}
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					if err := alerter.Alert(ctx, Alert{t, status, event, 0}); err != nil {
						log.Printf("alert failed for %s: %v", t.Name, err)
					}
					cancel()
				}
				lastUp = &status.Up
				out <- status
				<-ticker.C
			}
		}()
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	response := StatusResponse{
		statuses, hostStatus,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func checkConnectivity() bool {
	conn, err := net.DialTimeout("tcp", "8.8.8.8:53", 5*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func pollConnectivity(alerter Alerter, interval time.Duration) {
	ticker := time.NewTicker(interval)
	t := Target{"", ""}
	s := Status{"", false, 0, time.Now()}
	wasDown := false
	var timeWentDown time.Time
	defer ticker.Stop()
	for {
		up := checkConnectivity()
		if !up && !wasDown {
			timeWentDown = time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := alerter.Alert(ctx, Alert{t, s, EventNetworkDown, 0}); err != nil {
				log.Printf("failed to send alert: %v", err)
			}
			cancel()
			wasDown = true
		}
		if up && wasDown {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := alerter.Alert(ctx, Alert{t, s, EventNetworkRecovery, time.Since(timeWentDown)}); err != nil {
				log.Printf("failed to send alert: %v", err)
			}
			cancel()
			wasDown = false
		}

		<-ticker.C
	}
}

func checkRebootRequired() bool {
	_, err := os.Stat("/var/run/reboot-required")
	return err == nil
}

func checkUpdatesAvailable() (int, error) {
	cmd := exec.Command("apt", "list", "--upgradable")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	return len(strings.Split(strings.TrimSpace(string(out)), "\n")) - 1, nil
}

func main() {
	targets := []Target{
		{"Google", "https://www.google.com"},
		{"GitHub", "https://www.github.com"},
		{"1337", "http://localhost:5150/api/health"},
	}

	out := make(chan Status)

	alerter := DiscordAlerter{os.Getenv("DISCORD_WEBHOOK_URL")}
	monitor(targets, 10*time.Second, out, alerter)
	go pollConnectivity(alerter, 10*time.Second)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			updates, err := checkUpdatesAvailable()
			if err != nil {
				log.Printf("failed to check updates: %v", err)
			}

			mu.Lock()
			hostStatus = HostStatus{
				Connectivity:     checkConnectivity(),
				RebootRequired:   checkRebootRequired(),
				UpdatesAvailable: updates,
			}
			mu.Unlock()

			<-ticker.C
		}
	}()

	staticRoot, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("failed to load embedded static files: %v", err)
	}

	// 5152 sits next to 1337 (5150) and helm (5151); 8080 is qbittorrent's.
	// S3NTRY_ADDR overrides it if something else ever claims the port.
	addr := os.Getenv("S3NTRY_ADDR")
	if addr == "" {
		addr = ":5152"
	}

	http.HandleFunc("/status", statusHandler)
	http.Handle("/", http.FileServer(http.FS(staticRoot)))
	go func() {
		log.Printf("status page listening on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("HTTP server failed: %v", err)
		}
	}()

	for status := range out {
		mu.Lock()
		statuses[status.Target] = status
		mu.Unlock()
		log.Printf(
			"%s: up=%t latency=%dms checked %s",
			status.Target,
			status.Up,
			status.LatencyMs,
			status.CheckedAt,
		)
	}
}
