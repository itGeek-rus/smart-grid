package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
)

const querySelectByID = `
SELECT id, external_id, zone, name, type, status, meta, created_at, updated_at
FROM devices WHERE id = $1;
`

const querySelectByExternalID = `
SELECT id, external_id, zone, name, type, status, meta, created_at, updated_at
FROM devices WHERE external_id = $1;
`

const queryListByZone = `
SELECT id, external_id, zone, name, type, status, meta, created_at, updated_at
FROM devices WHERE zone = $1 ORDER BY id
`

const queryListAll = `
SELECT id, external_id, zone, name, type, status, meta, created_at, updated_at
FROM devices ORDER BY id
`

const queryUpsert = `
INSERT INTO devices (id, external_id, zone, name, type, status, meta, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
ON CONFLICT (id) DO UPDATE SET
	external_id = EXCLUDED.external_id,
	zone = EXCLUDED.zone,
	name = EXCLUDED.name,
	type = EXCLUDED.type,
	status = EXCLUDED.status,
	meta = EXCLUDED.meta,
	updated_at = now()
`

type DeviceRepo struct {
	pool *pgxpool.Pool
}

func NewDeviceRepo(pool *pgxpool.Pool) *DeviceRepo {
	return &DeviceRepo{pool: pool}
}

func (r *DeviceRepo) GetByID(ctx context.Context, id string) (*domain.Device, error) {
	row := r.pool.QueryRow(ctx, querySelectByID, id)
	return scanDevice(row)
}

func (r *DeviceRepo) GetByExternalID(ctx context.Context, externalID string) (*domain.Device, error) {
	row := r.pool.QueryRow(ctx, querySelectByExternalID, externalID)
	return scanDevice(row)
}

func (r *DeviceRepo) ListByZone(ctx context.Context, zone string) ([]domain.Device, error) {
	rows, err := r.pool.Query(ctx, queryListByZone, zone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Device
	for rows.Next() {
		d, err := scanDeviceRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

func (r *DeviceRepo) ListAll(ctx context.Context) ([]domain.Device, error) {
	rows, err := r.pool.Query(ctx, queryListAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Device
	for rows.Next() {
		d, err := scanDeviceRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

func (r *DeviceRepo) Upsert(ctx context.Context, device domain.Device) error {
	meta, _ := json.Marshal(device.Meta)
	_, err := r.pool.Exec(ctx, queryUpsert, device.ID, device.ExternalID,
		device.Zone, device.Name, device.Type, device.Status, meta)
	return err
}

type scannable interface {
	Scan(dest ...any) error
}

func scanDevice(row scannable) (*domain.Device, error) {
	var d domain.Device
	var meta []byte
	if err := row.Scan(&d.ID, &d.ExternalID, &d.Zone, &d.Name,
		&d.Type, &d.Status, &meta, &d.CreatedAt, &d.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal(meta, &d.Meta)
	return &d, nil
}

func scanDeviceRows(rows pgx.Rows) (*domain.Device, error) {
	return scanDevice(rows)
}
