package processor_test

import (
	"testing"
	"time"

	"github.com/itGeek-rus/smart-grid.git/internal/service/processor"
)

func TestDetectAnomaly_Voltage(t *testing.T) {
	t.Parallel()

	res := processor.DetectAnomaly(180, 100, 100, 10)
	if !res.IsAnomaly {
		t.Fatalf("expected anomaly, got: %+v", res)
	}
	if res.AlertType != "voltage_out_of_range" {
		t.Fatalf("alert type = %s, want voltage_out_of_range", res.AlertType)
	}
	if res.AnomalyScore < 0.7 {
		t.Fatalf("score = %v, want >= 0.7", res.AnomalyScore)
	}
}

func TestDetectAnomaly_PowerSpike(t *testing.T) {
	t.Parallel()

	res := processor.DetectAnomaly(220, 500, 100, 10)
	if !res.IsAnomaly {
		t.Fatalf("expected anomaly, got: %+v", res)
	}
	if res.AlertType != "power_spike" {
		t.Fatalf("alert type = %s, want power_spike", res.AlertType)
	}
}

func TestDetectAnomaly_Normal(t *testing.T) {
	t.Parallel()

	res := processor.DetectAnomaly(220, 110, 100, 10)
	if res.IsAnomaly {
		t.Fatalf("expected no anomaly, got: %+v", res)
	}
}

func TestDetectAnomaly_NotEnoughSamples(t *testing.T) {
	t.Parallel()

	res := processor.DetectAnomaly(220, 500, 100, 2)
	if res.IsAnomaly {
		t.Fatalf("expected no anomaly with few samples, got: %+v", res)
	}
}

func TestBuildFiveMinAggregate(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, 8, 2, 1, 7, 30, 0, time.UTC)
	agg := processor.BuildFiveMinAggregate("dev-001", 220, 5, 100, 150, 10, at)

	wantStart := time.Date(2026, 8, 2, 1, 5, 0, 0, time.UTC)
	wantEnd := wantStart.Add(5 * time.Minute)

	if agg.DeviceID != "dev-001" {
		t.Fatalf("device = %s", agg.DeviceID)
	}
	if !agg.WindowStart.Equal(wantStart) || !agg.WindowEnd.Equal(wantEnd) {
		t.Fatalf("window = [%v, %v], want [%v, %v]", agg.WindowStart, agg.WindowEnd, wantStart, wantEnd)
	}
	if agg.Samples != 10 || agg.MaxPower != 150 {
		t.Fatalf("unexpected aggregate: %+v", agg)
	}
}
