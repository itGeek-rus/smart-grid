package timescaledb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
)

const insertTelemetryQuery = `
INSERT INTO telemetry (time, device_id, zone, voltage, current, power, frequency,
                       received_at, payload)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

const listDeviceQuery = `
SELECT time, device_id, zone, voltage, current, power, frequency, received_at, payload
FROM telemetry
WHERE device_id = $1 AND time >= $2 AND time < $3
ORDER BY time DESC
LIMIT $4
`

const insertAggregateQuery = `
INSERT INTO telemetry_aggregates_5m
	(window_start, window_end, device_id, avg_voltage, avg_current, avg_power, max_power, samples)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (device_id, window_start) DO UPDATE SET
	window_end = EXCLUDED.window_end,
	avg_voltage = EXCLUDED.avg_voltage,
	avg_current = EXCLUDED.avg_current,
	avg_power = EXCLUDED.avg_power,
	max_power = EXCLUDED.max_power,
	samples = EXCLUDED.samples
`

type TelemetryRepo struct {
	pool *pgxpool.Pool
}

func NewTelemetryRepo(pool *pgxpool.Pool) *TelemetryRepo {
	return &TelemetryRepo{pool: pool}
}

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgx ping: %w", err)
	}
	return pool, nil
}

func (r *TelemetryRepo) Insert(ctx context.Context, t domain.Telemetry) error {
	payload, err := json.Marshal(t.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	_, err = r.pool.Exec(ctx, insertTelemetryQuery,
		t.MeasuredAt, t.DeviceID, t.Zone, t.Voltage,
		t.Current, t.Power, t.Frequency, t.ReceivedAt, payload)
	if err != nil {
		return fmt.Errorf("insert telemetry: %w", err)
	}
	return nil
}

func (r *TelemetryRepo) InsertBatch(ctx context.Context, items []domain.Telemetry) error {
	for _, item := range items {
		if err := r.Insert(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func (r *TelemetryRepo) ListByDevice(
	ctx context.Context,
	deviceID string,
	from, to time.Time,
	limit int,
) ([]domain.Telemetry, error) {
	rows, err := r.pool.Query(ctx, listDeviceQuery, deviceID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("list telemetry: %w", err)
	}
	defer rows.Close()

	var out []domain.Telemetry
	for rows.Next() {
		var t domain.Telemetry
		var payload []byte
		if err := rows.Scan(&t.MeasuredAt, &t.DeviceID, &t.Zone,
			&t.Voltage, &t.Current, &t.Power, &t.Frequency, &t.ReceivedAt, &payload); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(payload, &t.Payload)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TelemetryRepo) InsertAggregate(ctx context.Context, agg domain.TelemetryAggregate) error {
	_, err := r.pool.Exec(ctx, insertAggregateQuery,
		agg.WindowStart, agg.WindowEnd, agg.DeviceID, agg.AvgVoltage, agg.AvgCurrent, agg.AvgPower,
		agg.MaxPower, agg.Samples)
	if err != nil {
		return fmt.Errorf("insert aggregate: %w", err)
	}
	return nil
}
