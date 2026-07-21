package kafka

const (
	TopicRawTelemetry    = "raw.telemetry"
	TopicProcessedEvents = "processed.events"
	TopicAlerts          = "alerts"
	TopicCommand         = "commands"
	TopicRawTelemetryDLQ = "raw.telemetry.dlq"
)

var AllTopics = []string{
	TopicRawTelemetry,
	TopicProcessedEvents,
	TopicAlerts,
	TopicCommand,
	TopicRawTelemetryDLQ,
}
