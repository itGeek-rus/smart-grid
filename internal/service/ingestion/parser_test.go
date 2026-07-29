package ingestion_test

import (
	"errors"
	"testing"
	"time"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
	"github.com/itGeek-rus/smart-grid.git/internal/service/ingestion"
)

func TestParseAndValidate_OK(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"voltage":220.1,"current":5.2,"power":1144.5,"frequency":50,"measured_at":"2026-07-22T01:00:00Z"}`)
	receivedAt := time.Date(2026, 7, 22, 1, 5, 0, 0, time.UTC)

	evt, err := ingestion.ParseAndValidate("smartmeter/zone1/dev-001", raw, receivedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.DeviceID != "dev-001" || evt.Zone != "zone1" {
		t.Fatalf("unexpected ids: device=%s zone=%s", evt.DeviceID, evt.Zone)
	}
	if evt.EventType != domain.EventTypeRawTelemetry {
		t.Fatalf("event type = %s, want %s", evt.EventType, domain.EventTypeRawTelemetry)
	}
	if evt.EventID == "" {
		t.Fatal("expected non-empty event_id")
	}
	if !evt.OccurredAt.Equal(receivedAt) {
		t.Fatalf("occurred_at = %v, want %v", evt.OccurredAt, receivedAt)
	}
	if evt.Voltage != 220.1 || evt.Current != 5.2 || evt.Power != 1144.5 || evt.Frequency != 50 {
		t.Fatalf("unexpected metrics: %+v", evt)
	}
}

func TestParseAndValidate_DefaultFrequency(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"voltage":220,"current":1,"power":220,"measured_at":"2026-07-22T01:00:00Z"}`)
	evt, err := ingestion.ParseAndValidate("smartmeter/zone1/dev-001", raw, time.Now().UTC())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.Frequency != 50 {
		t.Fatalf("frequency = %v, want 50", evt.Frequency)
	}
}

func TestParseAndValidate_BadJSON(t *testing.T) {
	t.Parallel()

	_, err := ingestion.ParseAndValidate("smartmeter/zone1/dev-001", []byte(`{`), time.Now().UTC())
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func TestParseAndValidate_BadTopic(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"voltage":220,"current":1,"power":220,"measured_at":"2026-07-22T01:00:00Z"}`)
	_, err := ingestion.ParseAndValidate("bad-topic", raw, time.Now().UTC())
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func TestParseAndValidate_InvalidMetrics(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"voltage":-1,"current":1,"power":220,"measured_at":"2026-07-22T01:00:00Z"}`)
	_, err := ingestion.ParseAndValidate("smartmeter/zone1/dev-001", raw, time.Now().UTC())
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func TestParseTopic(t *testing.T) {
	t.Parallel()

	zone, deviceID, err := ingestion.ParseTopic("smartmeter/zone1/dev-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zone != "zone1" || deviceID != "dev-001" {
		t.Fatalf("got zone=%s device=%s", zone, deviceID)
	}
}
