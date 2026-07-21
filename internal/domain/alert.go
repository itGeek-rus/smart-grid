package domain

import "time"

type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

type AlertStatus string

const (
	AlertStatusOpen         AlertStatus = "open"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
)

type Alert struct {
	ID         string
	DeviceID   string
	Type       string
	Severity   AlertSeverity
	Status     AlertStatus
	Message    string
	Score      float64
	DetectedAt time.Time
	CreatedAt  time.Time
	Meta       map[string]string
}
