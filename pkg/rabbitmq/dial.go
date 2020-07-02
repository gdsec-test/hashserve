package rabbitmq

import (
	"github.com/streadway/amqp"
)

// Connection is a thin wrapper around amqp.Connection that stores state related to re-dialing.
type Connection struct {
	*amqp.Connection
}

// Dial creates a new AMQP Connection to the Broker located at uri.
func Dial(uri string) (*Connection, error) {
	conn := &Connection{}
	var err error

	conn.Connection, err = amqp.Dial(uri)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Channel creates a new custom Channel based on a prior Connection.
func (conn *Connection) Channel() (*Channel, error) {
	ch, err := conn.Connection.Channel()
	if err != nil {
		return nil, err
	}

	return &Channel{ch}, nil
}

