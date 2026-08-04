package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "smartgrid_http_requests_total",
		Help: "HTTP requests",
	}, []string{"service", "method", "path", "status"})

	HTTPDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "smartgrid_http_request_duration_seconds",
		Help:    "HTTP request duration",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method", "path"})

	MQTTMessages = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "smartgrid_mqtt_messages_total",
		Help: "MQTT messages handled",
	}, []string{"result"})

	KafkaConsumed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "smartgrid_kafka_consumed_total",
		Help: "Kafka messages consumed",
	}, []string{"topic", "result"})

	KafkaPublished = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "smartgrid_kafka_published_total",
		Help: "Kafka messages published",
	}, []string{"topic", "result"})

	Anomalies = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "smartgrid_anomalies_total",
		Help: "Detected anomalies",
	}, []string{"type", "severity"})

	TelemetryProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "smartgrid_telemetry_processed_total",
		Help: "Telemetry points processed",
	})
)

func Handler() http.Handler {
	return promhttp.Handler()
}
