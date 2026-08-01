package processor

import (
	"time"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
)

func BuildFiveMinAggregate(
	deviceID string,
	avgV, avgC, avgP, maxP float64,
	samples int64,
	at time.Time,
) domain.TelemetryAggregate {
	start := at.UTC().Truncate(5 * time.Minute)
	return domain.TelemetryAggregate{
		DeviceID:    deviceID,
		WindowStart: start,
		WindowEnd:   start.Add(5 * time.Minute),
		AvgVoltage:  avgV,
		AvgCurrent:  avgC,
		AvgPower:    avgP,
		MaxPower:    maxP,
		Samples:     samples,
	}
}
