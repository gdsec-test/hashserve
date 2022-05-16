package rabbitmq

import (
	"context"
	"time"

	"github.com/gdcorp-infosec/dcu-structured-logging-go/logger"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Producer struct {
	// The environment in which to run the application e.g. dev or prod
	// This variable is used by RabbitMQ to create the appropriate environment
	// specific namespacing for RabbitMQ Exchanges, Queues, and Bindings.
	env string

	// Producer amqp connection
	conn *Connection

	// Producer amqp channel
	ch *Channel

	confirms chan amqp.Confirmation
}

// NewProducer creates a new RabbitMQ Producer.
func NewProducer(ctx context.Context, env string, connection *Connection) (*Producer, error) {
	p := Producer{
		env:  env,
		conn: connection,
	}
	ch, err := p.conn.Channel()
	if err != nil {
		logger.Error(ctx, "failed to create channel", zap.Error(err))
		return nil, err
	}
	p.ch = ch
	p.confirms = p.ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	if err := p.ch.Confirm(false); err != nil {
		logger.Error(ctx, "Error in confirm", zap.Error(err))
		return nil, err
	}
	return &p, nil
}

func (p *Producer) Publish(ctx context.Context, messageContent []byte, exchangeName string) error {
	message := amqp.Publishing{
		Headers:      amqp.Table{},
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Time{},
		Body:         messageContent,
	}
	logger.Debug(ctx, "About to publish")
	for {
		err := p.ch.Publish(exchangeName,
			"#."+p.env+"-v2",
			false,
			false,
			message)
		confirmed := <-p.confirms
		if confirmed.Ack {
			break
		} else {
			logger.Error(ctx, "Publish failed", zap.Error(err))
			return err
		}
	}
	return nil
}
