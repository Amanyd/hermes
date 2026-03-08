package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type NatsPublisher struct {
	js nats.JetStreamContext
}

func NewNatsPublisher(url string) (*NatsPublisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream init: %w", err)
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "EVENTS",
		Subjects: []string{"events.*"},
	})
	if err != nil {
		log.Printf("EVENTS stream may already exist: %v", err)
	}
	return &NatsPublisher{js: js}, nil
}

type executionEvent struct {
	EventID    string         `json:"event_id"`
	RelayID    string         `json:"relay_id"`
	Payload    map[string]any `json:"payload"`
	ReceivedAt time.Time      `json:"received_at"`
}

func (p *NatsPublisher) PublishManualTrigger(_ context.Context, relayID string, payload map[string]any) error {
	event := executionEvent{
		EventID:    uuid.New().String(),
		RelayID:    relayID,
		Payload:    payload,
		ReceivedAt: time.Now().UTC(),
	}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	subject := fmt.Sprintf("events.%s", relayID)
	if _, err := p.js.Publish(subject, data); err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}
	return nil
}
