package rabbitmq

import "errors"

var (
	// ErrClosed is returned when an operation is attempted after Close.
	ErrClosed = errors.New("rabbitmq: client is closed")

	// ErrNack is returned by Producer.Publish when the broker negatively
	// acknowledges the message in confirm mode.
	ErrNack = errors.New("rabbitmq: broker nack")
)
