package rabbitmq

import (
	"os"

	"github.com/streadway/amqp"
)

// Channel serves as a simple wrapper around an amqp.Channel.
// It provides additional functionality for initializing the required RabbitMQ topology.
type Channel struct {
	*amqp.Channel
}

// Initialize will attempt to initialize all of the required exchanges, queues, and bindings
// that are necessary for a correct RabbitMQ topology.
//
// Finally, return a <-chan amqp.Delivery channel which serves as a way to consume messages
// that are produced on the queue above.
func (ch *Channel) Initialize(env string, prefetchCount int) (<-chan amqp.Delivery, error) {

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

func (ch *Channel) queueDeclareAndBind(env string) (amqp.Queue, error) {

	args := make(amqp.Table)
	queue_type := os.Getenv("queue-type")
	if queue_type == "quorum" {
		args["x-queue-type"] = queue_type
	} else {
		args = nil
	}

	q, err := ch.QueueDeclare(
		"hashserve-"+env, // Name
		true,             // Durable
		false,            // AutoDelete
		false,            // Exclusive
		false,            // NoWait
		args,             // Args
	)
	if err != nil {
		return amqp.Queue{}, err
	}

	err = ch.QueueBind(
		q.Name,      // Name
		"#."+env,    // Key
		"hashserve", // Exchange
		false,       // NoWait
		nil,         // Args
	)

	if err != nil {
		return amqp.Queue{}, err
	}

	return q, nil
}
