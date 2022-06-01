package rabbitmq

import (
	"github.com/streadway/amqp"
	"math/rand"
	"strings"
	"time"
)

// Connection is a thin wrapper around amqp.Connection that stores state related to re-dialing.
type Connection struct {
	*amqp.Connection
}

// Dial creates a new AMQP Connection to the Broker located at uri.
func Dial(uri string) (*Connection, error) {
	conn := &Connection{}
	var err error
	urls := strings.Split(uri, ";")
	//shuffles list of urls from https://yourbasic.org/golang/shuffle-slice-array/
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(urls), func(i, j int) { urls[i], urls[j] = urls[j], urls[i] })

	for _, url := range urls{
		conn.Connection, err = amqp.Dial(url)
		if err == nil {
			return conn, nil
		}
	}
		return nil, err

}

// Channel creates a new custom Channel based on a prior Connection.
func (conn *Connection) Channel() (*Channel, error) {
	ch, err := conn.Connection.Channel()
	if err != nil {
		return nil, err
	}

	return &Channel{ch}, nil
}

