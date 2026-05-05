package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Message is a delivered AMQP message with Ack/Nack helpers.
type Message struct {
	Body          []byte
	ContentType   string
	Headers       amqp.Table
	MessageID     string
	CorrelationID string
	RoutingKey    string
	Exchange      string
	Redelivered   bool

	// Ack acknowledges the message.
	Ack func() error
	// Nack rejects the message. When requeue is false, the broker forwards
	// the message to the dead-letter exchange configured on the queue (if
	// any) before discarding it.
	Nack func(requeue bool) error
}

// ConsumerConfig configures a Consumer.
type ConsumerConfig struct {
	Queue    string
	Tag      string
	Prefetch int

	// Logger overrides the Client logger for this Consumer. Optional.
	Logger *zap.Logger
}

func (c *ConsumerConfig) withDefaults() {
	if c.Prefetch <= 0 {
		c.Prefetch = 10
	}
}

// Consumer reads messages from a single queue on its own AMQP channel and
// re-subscribes automatically after a channel close or a connection drop.
//
// Dead-lettering is delegated to RabbitMQ: declare the queue with the
// "x-dead-letter-exchange" argument and Nack(requeue=false) will forward to
// the DLX without any additional code.
type Consumer struct {
	client *Client
	cfg    ConsumerConfig
	log    *zap.Logger
}

// NewConsumer creates a Consumer bound to the given client.
func NewConsumer(client *Client, cfg ConsumerConfig) *Consumer {
	cfg.withDefaults()
	log := cfg.Logger
	if log == nil {
		log = client.Logger()
	}
	return &Consumer{client: client, cfg: cfg, log: log}
}

// Consume returns a channel of Messages. The channel is closed when ctx is
// cancelled or the Client is closed. The caller MUST Ack or Nack every
// received Message.
func (c *Consumer) Consume(ctx context.Context) (<-chan Message, error) {
	ch, deliveries, err := c.subscribe(ctx)
	if err != nil {
		return nil, err
	}

	out := make(chan Message)
	go c.run(ctx, out, ch, deliveries)
	return out, nil
}

func (c *Consumer) run(
	ctx context.Context,
	out chan<- Message,
	ch *amqp.Channel,
	deliveries <-chan amqp.Delivery,
) {
	defer close(out)
	defer func() {
		if ch != nil {
			_ = ch.Close()
		}
	}()

	for {
		c.drain(ctx, deliveries, out)

		if ctx.Err() != nil {
			return
		}

		// Deliveries closed -> channel is gone. Re-subscribe with back-off.
		_ = ch.Close()
		ch = nil

		var (
			newCh  *amqp.Channel
			newDel <-chan amqp.Delivery
			err    error
		)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(c.client.ReconnectDelay()):
			}
			newCh, newDel, err = c.subscribe(ctx)
			if err == nil {
				break
			}
			c.log.Warn("rabbitmq: re-subscribe failed",
				zap.String("queue", c.cfg.Queue),
				zap.Error(err),
			)
		}
		ch, deliveries = newCh, newDel
	}
}

func (c *Consumer) subscribe(ctx context.Context) (*amqp.Channel, <-chan amqp.Delivery, error) {
	ch, err := c.client.Channel(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := ch.Qos(c.cfg.Prefetch, 0, false); err != nil {
		_ = ch.Close()
		return nil, nil, fmt.Errorf("rabbitmq: qos: %w", err)
	}
	deliveries, err := ch.Consume(
		c.cfg.Queue, c.cfg.Tag,
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return nil, nil, fmt.Errorf("rabbitmq: consume %q: %w", c.cfg.Queue, err)
	}
	return ch, deliveries, nil
}

func (c *Consumer) drain(ctx context.Context, deliveries <-chan amqp.Delivery, out chan<- Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}
			msg := wrap(d)
			select {
			case out <- msg:
			case <-ctx.Done():
				// Shutting down with an unprocessed message: requeue it.
				_ = d.Nack(false, true)
				return
			}
		}
	}
}

func wrap(d amqp.Delivery) Message {
	return Message{
		Body:          d.Body,
		ContentType:   d.ContentType,
		Headers:       d.Headers,
		MessageID:     d.MessageId,
		CorrelationID: d.CorrelationId,
		RoutingKey:    d.RoutingKey,
		Exchange:      d.Exchange,
		Redelivered:   d.Redelivered,
		Ack:           func() error { return d.Ack(false) },
		Nack:          func(requeue bool) error { return d.Nack(false, requeue) },
	}
}
