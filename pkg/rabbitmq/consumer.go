package rabbitmq

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gdcorp-infosec/dcu-structured-logging-go/logger"
	"github.com/streadway/amqp"
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

	// Number of image worker go routines
	nImageThreads int

	// Max retry count
	maxRetrycount int
}

// NewConsumer creates a new RabbitMQ Consumer.
func NewConsumer(env string, rmqURI string, nImageThreads int, maxRetrycount int) *Consumer {
	return &Consumer{
		env:           env,
		uri:           rmqURI,
		nImageThreads: nImageThreads,
		maxRetrycount: maxRetrycount,
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
	defer cancel()
	logger.Info(ctx, "connecting to amqp URI: ", zap.String("URI", c.uri))
	logger.Info(ctx, "connecting to amqp in ENV: ", zap.String("ENV", c.env))
	conn, err := Dial(c.uri, parentCtx)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	deliveries, err := ch.Initialize(c.env, c.nImageThreads*2)
	if err != nil {
		return err
	}

	// Handle sigterm signal
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	//Initialize the worker pool with all required channels. New amqp messages are fed to the jobschan, which distributes the job appropriately to image, video or text chan.
	worker := Worker{
		imageIngestChan: make(chan amqp.Delivery, c.nImageThreads),
		videoIngestChan: make(chan amqp.Delivery, c.nImageThreads),
		miscIngestChan:  make(chan amqp.Delivery, c.nImageThreads),
		jobsChan:        make(chan amqp.Delivery, c.nImageThreads),
		ctx:             ctx,
		cancelFunc:      cancel,
		env:             c.env,
		uri:             c.uri,
		conn:            conn,
		maxRetryCount:   c.maxRetrycount,
	}
	wg := &sync.WaitGroup{}
	// a single go routine for image and misc content and twice the number of
	//image threads for image worker and content type detection worker
	wg.Add(3 + c.nImageThreads)
	go worker.videoWorkerFunc(wg)
	go worker.miscWorkerFunc(wg)
	go worker.contentTypeWorker(wg)
	for iter := 0; iter < c.nImageThreads; iter++ {
		go worker.imageWorkerFunc(wg)
	}
	// Wait for hasher and hasher pdna before consuming messages
	for {
		respHasher, errHasher := http.Get("http://127.0.0.1:8080/health")
		if errHasher != nil || respHasher.StatusCode != 200 {
			logger.Info(ctx, "Hasher service is not up, sleeping for 5 seconds")
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
	logger.Info(ctx, "Consuming from rabbitmq")
	for {
		select {
		case <-termChan:
			logger.Info(ctx, "SIGINT signal caught")
			ch.Close()
			close(worker.imageIngestChan)
			close(worker.videoIngestChan)
			close(worker.miscIngestChan)
			close(worker.jobsChan)
			wg.Wait()
			logger.Info(ctx, "Workers exited gracefully")
			return nil
		case <-ctx.Done():
			logger.Info(ctx, "Done signal caught")
			close(worker.imageIngestChan)
			close(worker.videoIngestChan)
			close(worker.miscIngestChan)
			close(worker.jobsChan)
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
