package rabbitmq

import (
	"context"
	"fmt"
)

// Consumer abstracts the RabbitMQ connection and consumer loop from the caller.
// It reads messages from a delivery channel, adds them to a bundler, and executes
// the bundler's handler after a certain amount of time or size threshold is reached.
type Consumer struct {
	// The environment in which to run the application e.g. dev or prod
	// This variable is used by RabbitMQ to create the appropriate environment
	// specific namespacing for RabbitMQ Exchanges, Queues, and Bindings.
	env string

	// The complete AMQP broker URL to connect to complete with usernames,
	// passwords, ports, and virtual hosts.
	uri string

}

// NewConsumer creates a new RabbitMQ Consumer.
func NewConsumer(env, rmqURI string) *Consumer {
	return &Consumer{
		env:     env,
		uri:     rmqURI,
	}
}

// Serve creates a new Connection and opens a new Channel to a RabbitMQ Broker.
func (c *Consumer) Serve(ctx context.Context) error {
	conn, err := Dial(c.uri)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	deliveries, err := ch.Initialize(c.env)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-deliveries:
			// TODO: Add Hashing function calls here
			fmt.Println(msg.Body)
		}
	}
}
