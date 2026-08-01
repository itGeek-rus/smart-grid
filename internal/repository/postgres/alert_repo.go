package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
)

const queryCreateAlert = `
INSERT INTO alerts (id, device_id, type, severity, status, message, score, detected_at, created_at, meta)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`

const queryGetAlert = `
SELECT id, device_id, type, severity, status, message, score, detected_at, created_at, meta
FROM alerts WHERE id = $1
`

const queryListAlerts = `
SELECT id, device_id, type, severity, status, message, score, detected_at, created_at, meta
FROM alerts WHERE device_id = $1 AND status = 'open'
`

const queryUpdateAlert = `
UPDATE alerts SET status = $2 WHERE id = $1
`

type AlertRepo struct {
	pool *pgxpool.Pool
}

func NewAlertRepo(pool *pgxpool.Pool) *AlertRepo {
	return &AlertRepo{pool: pool}
}

func (r *AlertRepo) Create(ctx context.Context, alert domain.Alert) error {
	meta, _ := json.Marshal(alert.Meta)
	_, err := r.pool.Exec(ctx, queryCreateAlert,
		alert.ID, alert.DeviceID, alert.Type, alert.Severity, alert.Status,
		alert.Message, alert.Score, alert.DetectedAt, alert.CreatedAt, meta,
	)
	if err != nil {
		return fmt.Errorf("create alert: %w", err)
	}
	return nil
}

func (r *AlertRepo) GetByID(ctx context.Context, id string) (*domain.Alert, error) {
	row := r.pool.QueryRow(ctx, queryGetAlert, id)
	var a domain.Alert
	var meta []byte
	if err := row.Scan(&a.ID, &a.DeviceID, &a.Type, &a.Severity,
		&a.Status, &a.Message, &a.Score, &a.DetectedAt, &a.CreatedAt, &meta); err != nil {
		if errorsIsNoRows(err) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal(meta, &a.Meta)
	return &a, nil
}

func (r *AlertRepo) ListOpenByDevice(ctx context.Context, deviceID string) ([]domain.Alert, error) {
	rows, err := r.pool.Query(ctx, queryListAlerts, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Alert
	for rows.Next() {
		var a domain.Alert
		var meta []byte
		if err := rows.Scan(&a.ID, &a.DeviceID, &a.Type, &a.Severity,
			&a.Status, &a.Message, &a.Score, &a.DetectedAt, &a.CreatedAt, &meta); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(meta, &a.Meta)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AlertRepo) UpdateStatus(ctx context.Context, id string, status domain.AlertStatus) error {
	ct, err := r.pool.Exec(ctx, queryUpdateAlert, id, status)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func errorsIsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
