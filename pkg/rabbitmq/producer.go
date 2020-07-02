package rabbitmq

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"github.secureserver.net/digital-crimes/hashserve/pkg/types"
	"log"
	"time"
)

type Producer struct {
	// The environment in which to run the application e.g. dev or prod
	// This variable is used by RabbitMQ to create the appropriate environment
	// specific namespacing for RabbitMQ Exchanges, Queues, and Bindings.
	env string

	// The complete AMQP broker URL to connect to complete with usernames,
	// passwords, ports, and virtual hosts.
	uri string

}

// NewProducer creates a new RabbitMQ Producer.
func NewProducer(env, rmqURI string) *Producer {
	return &Producer{
		env:     env,
		uri:     rmqURI,
	}
}

func(p *Producer) Publish(tr *types.ThornWorkerRequest) error {
	conn, err := Dial(p.uri)
	if err != nil {
		log.Printf("failed to connect %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("failed to create channel %s", err)
	}
	defer ch.Close()

	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	if err := ch.Confirm(false); err != nil {
		log.Printf("confirmation destination %s", err)
	}

	json, err := json.Marshal(tr)
	if err != nil {
		log.Printf("unable to marshal message %s", err)
		return err
	}

	message := amqp.Publishing{
		Headers: amqp.Table{},
		ContentType: "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp: time.Time{},
		Body: json,
	}

	for {
		err = ch.Publish("pdna-processor",
			"#."+p.env,
			false,
			false,
			message)
		confirmed := <- confirms
		if confirmed.Ack {
			break
		} else {
			log.Printf("published failed")
			return err
		}
	}
	return nil
}
