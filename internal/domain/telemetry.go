package domain

import "time"

type Telemetry struct {
	ID         string
	DeviceID   string
	Zone       string
	Voltage    float64
	Current    float64
	Power      float64
	Frequency  float64
	MeasuredAt time.Time
	ReceivedAt time.Time
	Payload    map[string]any
}

type TelemetryAggregate struct {
	DeviceID    string
	WindowStart time.Time
	WindowEnd   time.Time
	AvgVoltage  float64
	AvgCurrent  float64
	AvgPower    float64
	MaxPower    float64
	Samples     int64
}
