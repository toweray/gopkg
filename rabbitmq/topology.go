package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ExchangeDecl describes an exchange to be declared.
type ExchangeDecl struct {
	Name       string
	Kind       string // direct | topic | fanout | headers
	Durable    bool
	AutoDelete bool
	Internal   bool
	Args       amqp.Table
}

// QueueDecl describes a queue to be declared. Set Args["x-dead-letter-exchange"]
// (and optionally "x-dead-letter-routing-key") to enable dead-lettering.
type QueueDecl struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	Args       amqp.Table
}

// BindingDecl describes a queue-to-exchange binding.
type BindingDecl struct {
	Queue      string
	Exchange   string
	RoutingKey string
	Args       amqp.Table
}

// Topology is a declarative description of the AMQP topology a service needs.
// Apply is idempotent (RabbitMQ declarations are no-ops if the entity already
// exists with matching properties), so it is safe to call on every startup.
type Topology struct {
	Exchanges []ExchangeDecl
	Queues    []QueueDecl
	Bindings  []BindingDecl
}

// Apply declares the topology on a fresh AMQP channel obtained from c.
func (t Topology) Apply(ctx context.Context, c *Client) error {
	ch, err := c.Channel(ctx)
	if err != nil {
		return err
	}
	defer ch.Close()

	for _, e := range t.Exchanges {
		if err := ch.ExchangeDeclare(
			e.Name, e.Kind, e.Durable, e.AutoDelete, e.Internal, false, e.Args,
		); err != nil {
			return fmt.Errorf("rabbitmq: declare exchange %q: %w", e.Name, err)
		}
	}
	for _, q := range t.Queues {
		if _, err := ch.QueueDeclare(
			q.Name, q.Durable, q.AutoDelete, q.Exclusive, false, q.Args,
		); err != nil {
			return fmt.Errorf("rabbitmq: declare queue %q: %w", q.Name, err)
		}
	}
	for _, b := range t.Bindings {
		if err := ch.QueueBind(
			b.Queue, b.RoutingKey, b.Exchange, false, b.Args,
		); err != nil {
			return fmt.Errorf(
				"rabbitmq: bind %q -> %q (%q): %w",
				b.Queue, b.Exchange, b.RoutingKey, err,
			)
		}
	}
	return nil
}
