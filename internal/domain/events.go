package domain

import "time"

const (
	EventTypeRawTelemetry       = "telemetry.raw"
	EventTypeProcessedTelemetry = "telemetry.processed"
	EventTypeAlertRaised        = "alert.raised"
	EventDTypeDeviceCommand     = "device.command"
)

type Envelope struct {
	EventID       string            `json:"event_id"`
	EventType     string            `json:"event_type"`
	SchemaVersion int               `json:"schema_version"`
	OccurredAt    time.Time         `json:"occurred_at"`
	TraceID       string            `json:"trace_id,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
}

type RawTelemetryEvent struct {
	Envelope
	DeviceID   string         `json:"device_id"`
	Zone       string         `json:"zone"`
	Voltage    float64        `json:"voltage"`
	Current    float64        `json:"current"`
	Power      float64        `json:"power"`
	Frequency  float64        `json:"frequency"`
	MeasuredAt time.Time      `json:"measured_at"`
	ReceivedAt time.Time      `json:"received_at"`
	Payload    map[string]any `json:"payload,omitempty"`
}

type ProcessedTelemetryEvent struct {
	Envelope
	DeviceID     string    `json:"device_id"`
	Zone         string    `json:"zone"`
	Power        float64   `json:"power"`
	AnomalyScore float64   `json:"anomaly_score"`
	ProcessedAt  time.Time `json:"processed_at"`
}

type AlertEvent struct {
	Envelope
	AlertID    string        `json:"alert_id"`
	DeviceID   string        `json:"device_id"`
	Type       string        `json:"type"`
	Severity   AlertSeverity `json:"severity"`
	Message    string        `json:"message"`
	Score      float64       `json:"score"`
	DetectedAt time.Time     `json:"detected_at"`
}

type DeviceCommandEvent struct {
	Envelope
	CommandID string            `json:"command_id"`
	DeviceID  string            `json:"device_id"`
	Command   string            `json:"command"`
	Params    map[string]string `json:"params,omitempty"`
	IssuedAt  time.Time         `json:"issued_at"`
}

type DLQEvent struct {
	Envelope
	SourceTopic string `json:"source_topic"`
	Reason      string `json:"reason"`
	RawPayload  string `json:"raw_payload"`
}
