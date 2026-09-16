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
