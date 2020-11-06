package rabbitmq

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/streadway/amqp"
	"github.secureserver.net/digital-crimes/hashserve/pkg/logger"
	"go.uber.org/zap"
)

// Consumer abstracts the RabbitMQ connection and consumer loop from the caller.
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
		env: env,
		uri: rmqURI,
	}
}

// Serve creates a new Connection and opens a new Channel to a RabbitMQ Broker.
/*
Over view of functionality:
1. Serve creates an amqp consumer and listens to sigint signal.
2. Serve also starts 4 additional go routines.
3. StartWorker go routine listens to amqp messages passed to the jobs chan by serve,
detects the content type and routes it to one of image ingest channel,
video ingest channel or miscellaneous ingest channel.
4. imageWorkerFunc, videoWorkerFunc and miscWorkerFunc go routines listens to the appropriate channel,
executes content type specific logic and publishes to its respective rabbitmq queue
*/
func (c *Consumer) Serve(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)
	logger.Info(ctx, "connecting to amqp URI: ", zap.String("URI", c.uri))
	logger.Info(ctx, "connecting to amqp in ENV: ", zap.String("ENV", c.env))
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

	// Handle sigterm signal
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	//Initialize the worker pool with all required channels. New amqp messages are fed to the jobschan, which distributes the job appropriately to image, video or text chan.
	worker := Worker{
		imageIngestChan: make(chan amqp.Delivery),
		videoIngestChan: make(chan amqp.Delivery),
		miscIngestChan:  make(chan amqp.Delivery),
		jobsChan:        make(chan amqp.Delivery),
		ctx:             ctx,
		cancelFunc:      cancel,
		env:             c.env,
		uri:             c.uri,
		conn:            conn,
	}
	wg := &sync.WaitGroup{}
	go worker.startWorker()
	wg.Add(3)
	go worker.imageWorkerFunc(wg)
	go worker.videoWorkerFunc(wg)
	go worker.miscWorkerFunc(wg)
	for {
		select {
		case <-termChan:
			logger.Info(ctx, "SIGINT signal caught")
			ch.Close()
			cancel()
			wg.Wait()
			logger.Info(ctx, "Workers exited gracefully")
			return nil
		case <-ctx.Done():
			logger.Info(ctx, "Done signal caught")
			ch.Close()
			wg.Wait()
			logger.Info(ctx, "Workers exited gracefully")
			return nil
		case msg := <-deliveries:
			logger.Debug(ctx, "Message received")
			worker.jobsChan <- msg
		}
	}
}
