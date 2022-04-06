package rabbitmq

import (
	"github.com/streadway/amqp"
	"os"
)

// Channel serves as a simple wrapper around an amqp.Channel.
// It provides additional functionality for initializing the required RabbitMQ topology.
type Channel struct {
	*amqp.Channel
}

var products = [...]string{
	"hashserve",
}

// Initialize will attempt to initialize all of the required exchanges, queues, and bindings
// that are necessary for a correct RabbitMQ topology.
//
// First it attempts to declare all of the product topic exchanges and bind all of these
// to an internal topic exchange used for routing e.g. product-binding -> thornworker-<env>
//
// Next, declare our consumer queue and bind it to thornworker-<env> topic exchange.
//
// Finally, return a <-chan amqp.Delivery channel which serves as a way to consume messages
// that are produced on the queue above.
func (ch *Channel) Initialize(env string, prefetchCount int) (<-chan amqp.Delivery, error) {
	if err := ch.exchangeDeclare(env); err != nil {
		return nil, err
	}
	if err := ch.exchangeBind(env); err != nil {
		return nil, err
	}

	q, err := ch.queueDeclareAndBind(env)
	if err != nil {
		return nil, err
	}

	// QoS determines how many RabbitMQ messages can be consumed for a given connection.
	// This prevents the main worker loop from pulling an unbounded amount of tasks despite
	// much of the downstream work being blocked by Semaphores and other mechanisms.
	if err := ch.Qos(prefetchCount, 0, false); err != nil {
		return nil, err
	}

	return ch.Consume(
		q.Name, // Queue
		"",     // Consumer
		false,  // AutoAck
		false,  // Exclusive
		false,  // NoLocal
		false,  // NoWait
		nil,    // Args
	)
}

func (ch *Channel) exchangeDeclare(env string) error {
	err := ch.ExchangeDeclare(
		"hashserve-"+env,   // Name
		amqp.ExchangeTopic, // Kind
		true,               // Durable
		false,              // AutoDelete
		true,               // Internal
		false,              // NoWait
		nil,                // Args
	)
	if err != nil {
		return err
	}

	for _, p := range products {
		if err := ch.ExchangeDeclare(
			p,                  // Name
			amqp.ExchangeTopic, // Kind
			true,               // Durable
			false,              // AutoDelete
			false,              // Internal
			false,              // NoWait
			nil,                // Args
		); err != nil {
			return err
		}
	}

	return nil
}

func (ch *Channel) exchangeBind(env string) error {
	for _, p := range products {
		if err := ch.ExchangeBind(
			"hashserve-"+env, // Destination
			"#."+env,         // Key
			p,                // Source
			false,            // NoWait
			nil,              // Args
		); err != nil {
			return err
		}
	}

	return nil
}


func (ch *Channel) queueDeclareAndBind(env string) (amqp.Queue, error) {

	args := make(amqp.Table)
	args["x-queue-type"] = os.Getenv("queue-type")

	q, err := ch.QueueDeclare(
		"hashserve-"+env, // Name
		true,             // Durable
		false,            // AutoDelete
		false,            // Exclusive
		false,            // NoWait
		args,              // Args
	)
	if err != nil {
		return amqp.Queue{}, err
	}

	err = ch.QueueBind(
		q.Name,           // Name
		"#."+env,         // Key
		"hashserve-"+env, // Exchange
		false,            // NoWait
		nil,              // Args
	)

	if err != nil {
		return amqp.Queue{}, err
	}

	return q, nil
}
