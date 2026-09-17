package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type AlertEvent string

const (
	EventDown     AlertEvent = "down"
	EventRecovery AlertEvent = "recovery"
)

type Alert struct {
	Target Target
	Status Status
	Event  AlertEvent
}
type Alerter interface {
	Alert(ctx context.Context, a Alert) error
}

type DiscordAlerter struct {
	URL string // for now, Discord Webdhook
}

type SlackAlerter struct {
	URL string // for now, Discord Webdhook
}

func (d DiscordAlerter) Alert(ctx context.Context, a Alert) error {
	var message string

	switch a.Event {
	case EventDown:
		message = "🔴 **" + a.Target.Name + " is DOWN**"
	case EventRecovery:
		message = "🟢 **" + a.Target.Name + " has RECOOVERED**"
	default:
		return fmt.Errorf("unknown alert event: %q", a.Event)
	}
	payload, err := json.Marshal(struct {
		Content string `json:"content"`
	}{
		message,
	})
	if err != nil {
		return err
	}
	return postWebhook(ctx, d.URL, payload)
}

func (s SlackAlerter) Alert(ctx context.Context, a Alert) error {
	var message string

	switch a.Event {
	case EventDown:
		message = "🔴 **" + a.Target.Name + " is DOWN**"
	case EventRecovery:
		message = "🟢 **" + a.Target.Name + " has RECOOVERED**"
	default:
		return fmt.Errorf("unknown alert event: %q", a.Event)
	}
	payload, err := json.Marshal(struct {
		Text string `json:"text"`
	}{
		message,
	})
	if err != nil {
		return err
	}
	return postWebhook(ctx, s.URL, payload)
}

func postWebhook(ctx context.Context, url string, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %s", resp.Status)
	}

	return nil
}
