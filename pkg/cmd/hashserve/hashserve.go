package hashserve

import (
	"context"
	"github.com/gdcorp-infosec/dcu-structured-logging-go/logger"
	"github.com/gdcorp-infosec/hashserve/pkg/rabbitmq"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"strconv"
)

// config provides a central location for all application specific configuration.
type config struct {
	// The environment in which to run the application e.g. dev or prod
	// This variable may also be used for other things such as correctly
	// namespacing RabbitMQ Exchanges, Queues, and Bindings.
	env string

	// Broker host to connect and consume messages from.
	amqpBroker string

	// Number of Image worker go routines
	nImageThread string

	// Log level
	logLevel string

	// Max retry count
	maxRetryCount string
}

// load attempts to load all necessary environment variables needed to run the application.
// The absence of any of these environment variables will return an err, else nil.
func (w *config) load() (err error) {
	if err = w.loadEnv("ENV", &w.env); err != nil {
		return
	}


	if err = w.loadEnv("MULTIPLE_BROKERS", &w.amqpBroker); err != nil {
			return
		}
	if err = w.loadEnv("NO_IMAGE_WORKER_THREADS", &w.nImageThread); err != nil {
		return
	}
	if err = w.loadEnv("MAX_RETRY_COUNT", &w.maxRetryCount); err != nil {
		return
	}
	if err = w.loadEnv("LOG_LEVEL", &w.logLevel); err != nil {
		//Defaulting to info log level if log level is not present
		w.logLevel = "INFO"
		err = nil
	}
	return
}

// loadEnv attempts to look for an environment variable with name and
// loads it into the destination pointed to by dst. If the environment
// variable does not exist, it returns an error, else nil.
func (w *config) loadEnv(name string, dst *string) error {
	if *dst != `` {
		return nil // Variable is already set
	}
	v, ok := os.LookupEnv(name)
	if !ok {
		return errors.Errorf("env var %q not set", name)
	}

	*dst = v
	return nil
}

// Run initializes the baseline application, loggers, and other things necessary to Work.
func Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	config := config{}
	if err := config.load(); err != nil {
		return err
	}
	lr, lrUndo, err := logger.New(config.logLevel, "stderr")
	if err != nil {
		return err
	}
	defer lrUndo()
	ctx = logger.WithContext(ctx, lr)
	logger.Info(ctx, "app object created successfully")
	return Work(ctx, &config)
}

// Work serves as the main work function of the application.
//
// It is responsible for loading application specific configurations as well as
// serving the main work loop.
func Work(ctx context.Context, config *config) error {
	uri := config.amqpBroker

	nImageThreadInt, err := strconv.Atoi(config.nImageThread)
	if err != nil {
		logger.Error(ctx, "Unable to convert NO_IMAGE_WORKER_THREADS configuration to int")
		return err
	}
	maxRetryCountInt, err := strconv.Atoi(config.maxRetryCount)
	if err != nil {
		logger.Error(ctx, "Unable to convert MAX_RETRY_COUNT configuration to int")
		return err
	}
	w := rabbitmq.NewConsumer(config.env, uri, nImageThreadInt, maxRetryCountInt)
	err = w.Serve(ctx)
	if err != nil {
		logger.Error(ctx, "main: unable to perform work", zap.Error(err))
		return err
	}
	return err
}
