package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckOnce(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Healthy"))
	}))
	defer server.Close()

	t.Run("service is up", func(t *testing.T) {
		target := Target{
			"test-service",
			server.URL,
		}
		status := checkOnce(target)
		if !status.Up {
			t.Error("expected service to be up")
		}
	})

	t.Run("service is down", func(t *testing.T) {
		target := Target{
			"test-service",
			"http://127.0.0.1:1",
		}

		status := checkOnce(target)

		if status.Up {
			t.Error("expected service to be down")
		}
	})
}
