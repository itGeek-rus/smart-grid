package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
)

type Cache struct {
	client *goredis.Client
}

func NewCache(addr, password string, db int) *Cache {
	return &Cache{
		client: goredis.NewClient(&goredis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c *Cache) AddSample(ctx context.Context, deviceID string, power float64, at time.Time) error {
	key := "telemetry:window:" + deviceID
	member := fmt.Sprintf("%d:%f", at.UnixMilli(), power)
	pipe := c.client.TxPipeline()
	pipe.ZAdd(ctx, key, goredis.Z{
		Score:  float64(at.UnixMilli()),
		Member: member,
	})
	pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(at.Add(-10*time.Minute).UnixMilli(), 10))
	pipe.Expire(ctx, key, 15*time.Minute)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *Cache) WindowStats(
	ctx context.Context,
	deviceID string,
	window time.Duration,
) (avgPower, maxPower float64,
	samples int64,
	err error,
) {
	key := "telemetry:window:" + deviceID
	minScore := strconv.FormatInt(time.Now().UTC().Add(-window).UnixMilli(), 10)
	vals, err := c.client.ZRangeArgs(ctx, goredis.ZRangeArgs{
		Key:     key,
		Start:   minScore,
		Stop:    "+inf",
		ByScore: true,
	}).Result()
	if err != nil {
		return 0, 0, 0, err
	}
	if len(vals) == 0 {
		return 0, 0, 0, nil
	}
	var sum float64
	for _, v := range vals {
		var ts int64
		var p float64
		if _, scanErr := fmt.Sscanf(v, "%d:%f", &ts, &p); scanErr != nil {
			continue
		}
		sum += p
		if p > maxPower {
			maxPower = p
		}
		samples++
	}
	if samples > 0 {
		avgPower = sum / float64(samples)
	}
	return avgPower, maxPower, samples, nil
}

func (c *Cache) SetLastTelemetry(ctx context.Context, t domain.Telemetry) error {
	body, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "telemetry:last:"+t.DeviceID, body, 10*time.Minute).Err()
}

func (c *Cache) GetLastTelemetry(ctx context.Context, deviceID string) (*domain.Telemetry, error) {
	body, err := c.client.Get(ctx, "telemetry:last:"+deviceID).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var t domain.Telemetry
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
