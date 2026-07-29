package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/itGeek-rus/smart-grid.git/internal/repository"
)

type Client struct {
	client pahomqtt.Client
	log    *slog.Logger
	mu     sync.Mutex
}

func NewClient(brokerURL, clientID string, log *slog.Logger) (*Client, error) {
	opts := pahomqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(2 * time.Second).
		SetKeepAlive(30 * time.Second).
		SetOrderMatters(false)

	c := pahomqtt.NewClient(opts)
	token := c.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		return nil, fmt.Errorf("mqtt connection timeout")
	}
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("mqtt error: %w", err)
	}

	return &Client{client: c, log: log}, nil
}

func (c *Client) Subscribe(ctx context.Context, topic string, qos byte, handler repository.MQTTHandler) error {
	token := c.client.Subscribe(topic, qos, func(_ pahomqtt.Client, msg pahomqtt.Message) {
		if err := handler(ctx, repository.MQTTMessage{
			Topic:   msg.Topic(),
			Payload: append([]byte(nil), msg.Payload()...),
		}); err != nil {
			c.log.Error("mqtt handler failed",
				slog.String("topic", msg.Topic()),
				slog.String("error", err.Error()),
			)
		}
	})
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("mqtt subscribe timeout: %s", topic)
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("mqtt subscribe %s: %w", topic, err)
	}
	c.log.Info("mqtt subscribed", slog.String("topic", topic), slog.Int("qos", int(qos)))
	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(250)
	}
	return nil
}
