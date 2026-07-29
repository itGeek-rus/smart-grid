-- +goose Up
INSERT INTO devices (id, external_id, zone, name, type, status)
VALUES
    ('dev-001', 'meter-zone1-001', 'zone1', 'Smart Meter 001', 'smart_meter', 'active'),
    ('dev-002', 'meter-zone2-002', 'zone2', 'Smart Meter 002', 'smart_meter', 'active'),
    ('dev-003', 'meter-zone3-003', 'zone3', 'Smart Meter 003', 'smart_meter', 'active')
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM devices WHERE id IN ('dev-001', 'dev-002', 'dev-003');