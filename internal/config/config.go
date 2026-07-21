package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	MQTT     MQTTConfig
}

type AppConfig struct {
	Name            string        `env:"APP_NAME" envDefault:"smart-grid-processor"`
	Env             string        `env:"APP_ENV" envDefault:"local"`
	LogLevel        string        `env:"APP_LOG_LEVEL" envDefault:"info"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

type HTTPConfig struct {
	Addr string `env:"HTTP_ADDR" envDefault:":8080"`
}

type PostgresConfig struct {
	Host     string `env:"DB_HOST"`
	Port     int    `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Name     string `env:"DB_NAME"`
	SSLMode  string `env:"DB_SSLMODE"`
}

func (p PostgresConfig) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(p.User, p.Password),
		Host:   fmt.Sprintf("%s:%d", p.Host, p.Port),
		Path:   p.Name,
	}
	q := u.Query()
	q.Set("sslmode", p.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

type KafkaConfig struct {
	Brokers              []string `env:"KAFKA_BROKERS" envSeparator:"," envDefault:"localhost:9092"`
	ClientID             string   `env:"KAFKA_CLIENT_ID" envDefault:"smart-grid-processor"`
	TopicRawTelemetry    string   `env:"KAFKA_TOPIC_RAW_TELEMETRY" envDefault:"raw.telemetry"`
	TopicProcessedEvents string   `env:"KAFKA_TOPIC_PROCESSED_EVENTS" envDefault:"processed.events"`
	TopicAlerts          string   `env:"KAFKA_TOPIC_ALERTS" envDefault:"alerts"`
	TopicDLQ             string   `env:"KAFKA_TOPIC_DLQ" envDefault:"raw.telemetry.dlq"`
	ConsumeGroup         string   `env:"KAFKA_CONSUMER_GROUP" envDefault:"processor"`
}

type MQTTConfig struct {
	BrokerURL      string `env:"MQTT_BROKER_URL" envDefault:"localhost:1883"`
	ClientID       string `env:"MQTT_CLIENT_ID" envDefault:"smart-grid-ingestion"`
	TopicTelemetry string `env:"MQTT_TOPIC_TELEMETRY" envDefault:"smartmeter/+/+"`
	Qos            byte   `env:"MQTT_QOS" envDefault:"1"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func (c Config) IsLocal() bool {
	return c.App.Env == "local" || c.App.Env == "dev"
}
