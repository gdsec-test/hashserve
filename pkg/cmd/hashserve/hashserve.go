package hashserve

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.secureserver.net/digital-crimes/hashserve/pkg/rabbitmq"
	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.secureserver.net/digital-crimes/hashserve/pkg/logger"
)

// config provides a central location for all application specific configuration.
type config struct {
	// The environment in which to run the application e.g. dev or prod
	// This variable may also be used for other things such as correctly
	// namespacing RabbitMQ Exchanges, Queues, and Bindings.
	env string

	// Username to use when connecting to the AMQP Broker.
	amqpUser string

	// Password to use when connecting to the AMQP broker.
	amqpPassword string

	// Broker host to connect and consume messages from.
	amqpBroker string

	// Log level
	logLevel string
}

// load attempts to load all necessary environment variables needed to run the application.
// The absence of any of these environment variables will return an err, else nil.
func (w *config) load() (err error) {
	if err = w.loadEnv("ENV", &w.env); err != nil {
		return
	}
	if err = w.loadEnv("AMQP_USER", &w.amqpUser); err != nil {
		return
	}
	if err = w.loadEnv("AMQP_PASSWORD", &w.amqpPassword); err != nil {
		return
	}
	if err = w.loadEnv("AMQP_BROKER", &w.amqpBroker); err != nil {
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
	uri := fmt.Sprintf("amqps://%s:%s@%s:5672/pdna", config.amqpUser, url.QueryEscape(config.amqpPassword), config.amqpBroker)
	w := rabbitmq.NewConsumer(config.env, uri)
	err := w.Serve(ctx)
	if err != nil {
		logger.Error(ctx, "main: unable to perform work", zap.Error(err))
		return err
	}
	return err
}
