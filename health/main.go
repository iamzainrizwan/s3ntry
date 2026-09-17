package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

type Target struct {
	Name string
	URL  string // whatever HTTP health checkpoint
}

type Status struct {
	Target    string
	Up        bool
	LatencyMs int64
	CheckedAt time.Time
}

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

func monitor(targets []Target, interval time.Duration, out chan<- Status) {
	for _, target := range targets {
		t := target
		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			out <- checkOnce(t)
			for {
				<-ticker.C
				out <- checkOnce(t)
			}
		}()
	}
}

func main() {
	targets := []Target{
		{"Google", "https://www.google.com"},
		{"GitHub", "https://www.github.com"},
	}

	out := make(chan Status)

	monitor(targets, 10*time.Second, out)

	for status := range out {
		log.Printf(
			"%s: up=%t latency=%dms checked %s",
			status.Target,
			status.Up,
			status.LatencyMs,
			status.CheckedAt,
		)
	}
}
