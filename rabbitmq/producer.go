package rabbitmq

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// PublishOption mutates an amqp.Publishing before it is sent.
type PublishOption func(*amqp.Publishing)

func WithContentType(ct string) PublishOption {
	return func(p *amqp.Publishing) { p.ContentType = ct }
}

func WithMessageID(id string) PublishOption {
	return func(p *amqp.Publishing) { p.MessageId = id }
}

func WithCorrelationID(id string) PublishOption {
	return func(p *amqp.Publishing) { p.CorrelationId = id }
}

func WithHeaders(h amqp.Table) PublishOption {
	return func(p *amqp.Publishing) { p.Headers = h }
}

func WithExpiration(d time.Duration) PublishOption {
	return func(p *amqp.Publishing) {
		p.Expiration = strconv.FormatInt(d.Milliseconds(), 10)
	}
}

// ProducerConfig configures a Producer.
type ProducerConfig struct {
	// Mandatory marks every published message as mandatory. The broker
	// will return undeliverable messages, but Publish does not currently
	// surface returns; configure NotifyReturn on the channel directly if
	// needed.
	Mandatory bool

	// Logger overrides the Client logger for this Producer. Optional.
	Logger *zap.Logger
}

// Producer publishes messages with publisher confirms on its own dedicated
// AMQP channel. The channel is opened lazily and reopened transparently
// after a channel-level error or a connection drop.
//
// Producer is safe for concurrent use.
type Producer struct {
	client *Client
	cfg    ProducerConfig
	log    *zap.Logger

	mu sync.Mutex
	ch *amqp.Channel
}

// NewProducer wraps a Client with a Producer.
func NewProducer(client *Client, cfg ProducerConfig) *Producer {
	log := cfg.Logger
	if log == nil {
		log = client.Logger()
	}
	return &Producer{client: client, cfg: cfg, log: log}
}

// Publish sends body to (exchange, routingKey) and waits for the broker
// confirmation. Returns ErrNack on broker nack and ctx.Err() on cancellation.
func (p *Producer) Publish(
	ctx context.Context,
	exchange, routingKey string,
	body []byte,
	opts ...PublishOption,
) error {
	pub := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         body,
	}
	for _, o := range opts {
		o(&pub)
	}

	ch, err := p.acquireChannel(ctx)
	if err != nil {
		return err
	}

	confirm, err := ch.PublishWithDeferredConfirmWithContext(
		ctx, exchange, routingKey, p.cfg.Mandatory, false, pub,
	)
	if err != nil {
		// Channel is likely broken; drop it so the next call reopens.
		p.dropChannel()
		return fmt.Errorf("rabbitmq: publish: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-confirm.Done():
		if !confirm.Acked() {
			return ErrNack
		}
		return nil
	}
}

// Close releases the producer's channel. The underlying Client is not
// affected.
func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch == nil {
		return nil
	}
	err := p.ch.Close()
	p.ch = nil
	return err
}

// acquireChannel returns the producer's channel, opening (or reopening) it
// on demand and enabling confirm mode.
func (p *Producer) acquireChannel(ctx context.Context) (*amqp.Channel, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ch != nil && !p.ch.IsClosed() {
		return p.ch, nil
	}
	if p.ch != nil {
		_ = p.ch.Close()
		p.ch = nil
	}

	ch, err := p.client.Channel(ctx)
	if err != nil {
		return nil, err
	}
	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		return nil, fmt.Errorf("rabbitmq: enable confirms: %w", err)
	}

	p.ch = ch
	return ch, nil
}

func (p *Producer) dropChannel() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch != nil {
		_ = p.ch.Close()
		p.ch = nil
	}
}
