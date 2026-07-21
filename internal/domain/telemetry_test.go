package domain_test

import (
	"testing"
	"time"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
)

func TestRawTelemetryEvent_Schema(t *testing.T) {
	now := time.Now().UTC()
	evt := domain.RawTelemetryEvent{
		Envelope: domain.Envelope{
			EventID:       "evt-1",
			EventType:     domain.EventTypeRawTelemetry,
			SchemaVersion: 1,
			OccurredAt:    now,
		},
		DeviceID:   "dev-001",
		Zone:       "zone1",
		Voltage:    220.1,
		Current:    5.2,
		Power:      1144.5,
		Frequency:  50,
		MeasuredAt: now,
		ReceivedAt: now,
	}
	if evt.EventType != domain.EventTypeRawTelemetry {
		t.Fatalf("event type = %s", evt.EventType)
	}
	if evt.DeviceID == "" {
		t.Fatal("device_id required")
	}
}
