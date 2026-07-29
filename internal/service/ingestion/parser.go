package ingestion

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
)

type mqttPayload struct {
	Voltage    float64        `json:"voltage"`
	Current    float64        `json:"current"`
	Power      float64        `json:"power"`
	Frequency  float64        `json:"frequency"`
	MeasuredAt time.Time      `json:"measured_at"`
	Payload    map[string]any `json:"payload,omitempty"`
}

func ParseTopic(topic string) (zone, deviceID string, err error) {
	parts := strings.Split(topic, "/")
	if len(parts) != 3 || parts[0] != "smartmeter" || parts[1] == "" || parts[2] == "" {
		return "", "", fmt.Errorf("%w: topic %q", domain.ErrInvalidInput, topic)
	}
	return parts[1], parts[2], nil
}

func ParseAndValidate(topic string, raw []byte, receivedAt time.Time) (domain.RawTelemetryEvent, error) {
	zone, deviceID, err := ParseTopic(topic)
	if err != nil {
		return domain.RawTelemetryEvent{}, err
	}

	var p mqttPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return domain.RawTelemetryEvent{}, fmt.Errorf("%w: json: %w", domain.ErrInvalidInput, err)
	}

	if p.MeasuredAt.IsZero() {
		return domain.RawTelemetryEvent{}, fmt.Errorf("%w: measured_at required", domain.ErrInvalidInput)
	}
	if p.Frequency == 0 {
		p.Frequency = 50
	}
	if p.Voltage <= 0 || p.Current < 0 || p.Power < 0 {
		return domain.RawTelemetryEvent{}, fmt.Errorf("%w: invalid metrics", domain.ErrInvalidInput)
	}

	return domain.RawTelemetryEvent{
		Envelope: domain.Envelope{
			EventID:       uuid.NewString(),
			EventType:     domain.EventTypeRawTelemetry,
			SchemaVersion: 1,
			OccurredAt:    receivedAt.UTC(),
		},
		DeviceID:   deviceID,
		Zone:       zone,
		Voltage:    p.Voltage,
		Current:    p.Current,
		Power:      p.Power,
		Frequency:  p.Frequency,
		MeasuredAt: p.MeasuredAt.UTC(),
		ReceivedAt: receivedAt.UTC(),
		Payload:    p.Payload,
	}, nil
}
