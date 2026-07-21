-- +goose Up
CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS devices (
  id TEXT PRIMARY KEY,
  external_id TEXT NOT NULL UNIQUE,
  zone TEXT NOT NULL,
  name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  meta   JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_devices_zone ON devices (zone);

CREATE TABLE IF NOT EXISTS telemetry (
  time TIMESTAMPTZ NOT NULL,
  device_id TEXT NOT NULL REFERENCES device(id),
  zone TEXT NOT NULL,
  voltage DOUBLE PRECISION NOT NULL,
  current DOUBLE PRECISION NOT NULL,
  power DOUBLE PRECISION NOT NULL,
  frequency DOUBLE PRECISION NOT NULL DEFAULT 50,
  received_at TIMESTAMPTZ NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb
);

SELECT create_hypertable('telemetry', 'time', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_telemetry_device_time
    ON telemetry (device_id, time DESC);

CREATE TABLE IF NOT EXISTS telemetry_aggregates_5m (
    window_start TIMESTAMPTZ NOT NULL,
    window_end TIMESTAMPTZ NOT NULL,
    device_id TEXT NOT NULL REFERENCES device(id),
    avg_voltage DOUBLE PRECISION NOT NULL,
    avg_current DOUBLE PRECISION NOT NULL,
    avg_power DOUBLE PRECISION NOT NULL,
    max_power DOUBLE PRECISION NOT NULL,
    samples BIGINT NOT NULL,
    PRIMARY KEY(device_id, window_start)
);

SELECT create_hypertable('telemetry_aggregates_5m', 'window_start', if_not_exists => TRUE);

CREATE TABLE IF NOT EXISTS alerts (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL REFERENCES devices(id),
    type TEXT NOT NULL,
    severity TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    message TEXT NOT NULL,
    score DOUBLE PRECISION NOT NULL DEFAULT 0,
    detected_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    metaJSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_alerts_device_status
    ON alerts (device_id, status);

-- +goose Down
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS telemetry_aggrefates_5m;
DROP TABLE IF EXISTS telemetry;
DROP TABLE IF EXISTS devices;