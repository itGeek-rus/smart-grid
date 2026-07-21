package domain

import "time"

type DeviceStatus string

const (
	DeviceStatusActive   DeviceStatus = "active"
	DeviceStatusInactive DeviceStatus = "inactive"
	DeviceStatusFault    DeviceStatus = "fault"
)

type Device struct {
	ID         string
	ExternalID string
	Zone       string
	Name       string
	Type       string
	Status     DeviceStatus
	Meta       map[string]string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
