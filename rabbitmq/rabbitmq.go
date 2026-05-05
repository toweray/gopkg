package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const defaultReconnectDelay = 2 * time.Second

// Config configures a Client.
type Config struct {
	// URL is a RabbitMQ AMQP URL, e.g. amqp://user:pass@host:5672/vhost.
	URL string

	// AMQPConfig is forwarded to amqp.DialConfig. Optional.
	AMQPConfig amqp.Config

	// ReconnectDelay is the back-off between reconnect attempts.
	// If zero, defaults to 2s.
	ReconnectDelay time.Duration

	// Logger receives connection lifecycle events. If nil, logs are dropped.
	Logger *zap.Logger
}

func (c *Config) withDefaults() {
	if c.ReconnectDelay <= 0 {
		c.ReconnectDelay = defaultReconnectDelay
	}
	if c.Logger == nil {
		c.Logger = zap.NewNop()
	}
}

// Client manages a single long-lived RabbitMQ connection with automatic
// reconnection. Callers obtain a fresh AMQP channel through Channel() and
// must not share channels between goroutines.
type Client struct {
	cfg Config

	mu     sync.RWMutex
	conn   *amqp.Connection
	closed bool

	stopCh chan struct{}
	doneCh chan struct{}

	log *zap.Logger
}

// NewClient dials RabbitMQ once and starts a background supervisor that
// reconnects on connection drops. The first dial must succeed; subsequent
// failures are retried indefinitely until Close.
func NewClient(cfg Config) (*Client, error) {
	cfg.withDefaults()

	c := &Client{
		cfg:    cfg,
		log:    cfg.Logger,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}

	if err := c.connect(); err != nil {
		return nil, err
	}

	go c.supervise()
	return c, nil
}

// Channel opens a fresh AMQP channel from the current connection. If the
// connection is currently down, Channel waits up to ctx for it to come back.
func (c *Client) Channel(ctx context.Context) (*amqp.Channel, error) {
	for {
		c.mu.RLock()
		if c.closed {
			c.mu.RUnlock()
			return nil, ErrClosed
		}
		conn := c.conn
		c.mu.RUnlock()

		if conn != nil && !conn.IsClosed() {
			ch, err := conn.Channel()
			if err == nil {
				return ch, nil
			}
			c.log.Warn("rabbitmq: open channel failed", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-c.stopCh:
			return nil, ErrClosed
		case <-time.After(c.cfg.ReconnectDelay):
		}
	}
}

// ReconnectDelay returns the configured back-off between reconnects.
func (c *Client) ReconnectDelay() time.Duration {
	return c.cfg.ReconnectDelay
}

// Logger returns the configured logger.
func (c *Client) Logger() *zap.Logger {
	return c.log
}

// IsReady reports whether the client currently has a healthy connection.
func (c *Client) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.closed && c.conn != nil && !c.conn.IsClosed()
}

// Close shuts the client down: stops the supervisor, closes the connection,
// and unblocks any callers waiting in Channel. It is safe to call multiple
// times.
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()

	close(c.stopCh)
	<-c.doneCh

	if conn != nil {
		return conn.Close()
	}
	return nil
}

func (c *Client) connect() error {
	conn, err := amqp.DialConfig(c.cfg.URL, c.cfg.AMQPConfig)
	if err != nil {
		return fmt.Errorf("rabbitmq: dial: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	c.log.Info("rabbitmq: connected")
	return nil
}

// supervise watches the active connection and reconnects on close.
func (c *Client) supervise() {
	defer close(c.doneCh)

	for {
		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			if !c.waitDelay() {
				return
			}
			if err := c.connect(); err != nil {
				c.log.Error("rabbitmq: reconnect failed", zap.Error(err))
			}
			continue
		}

		notify := conn.NotifyClose(make(chan *amqp.Error, 1))
		select {
		case <-c.stopCh:
			return
		case err := <-notify:
			if err != nil {
				c.log.Warn("rabbitmq: connection lost", zap.Error(err))
			}
			c.mu.Lock()
			c.conn = nil
			c.mu.Unlock()
		}
	}
}

// waitDelay sleeps for ReconnectDelay or returns false on stop.
func (c *Client) waitDelay() bool {
	select {
	case <-c.stopCh:
		return false
	case <-time.After(c.cfg.ReconnectDelay):
		return true
	}
}
