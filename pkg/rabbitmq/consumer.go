package rabbitmq

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/streadway/amqp"
	"github.secureserver.net/digital-crimes/hashserve/pkg/logger"
	"github.secureserver.net/digital-crimes/hashserve/pkg/types"
	"go.uber.org/zap"
)

// Consumer abstracts the RabbitMQ connection and consumer loop from the caller.
// It reads messages from a delivery channel, adds them to a bundler, and executes
// the bundler's handler after a certain amount of time or size threshold is reached.
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
		env:     env,
		uri:     rmqURI,
	}
}

func ackMessage(ctx context.Context, msg amqp.Delivery) error{
	if objErr := msg.Ack(false); objErr != nil {
		logger.Error(ctx,"error acknowledging message", zap.Error(objErr))
		return objErr
	}
	return nil;
}

// Serve creates a new Connection and opens a new Channel to a RabbitMQ Broker.
func (c *Consumer) Serve(ctx context.Context) error {

	conn, err := Dial(c.uri)

	logger.Info(ctx, "connected to amqp URI: ",zap.String("URI", c.uri))
	logger.Info(ctx, "connected to amqp in ENV: ",zap.String("URI", c.env))

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

	var httpClient = &http.Client{}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-deliveries:
			/*  Basic Functionalities of the following goRoutine
			1) Read the published message by the hosting products
			2) Unmarshall the same into type ScanRequest
			3) Send the embedded URL (in the message body) over to hasher service
			4) Get the pDNA and MD5 Hash as response from the hasher service
			5) Creates and validate the ThornWorkerRequest
			6) Publish the new request to ThornWorker queue
			*/

			message := string(msg.Body)
			logger.Info(ctx, "Msg received",zap.String("msg", message))

			scanRequestData := types.ScanRequest{}
			err := json.Unmarshal([]byte(message), &scanRequestData)

			//If unable to unmarshal the message into scanRequestData, log the error.
			if err != nil {
				logger.Error(ctx, "failed to unmarshall json string into scanRequestData struct", zap.Error(err))
			} else {
				hashRequest := types.HashRequest{
					URL:  scanRequestData.URL,
				}

				err:= hashRequest.ValidateRequiredFields()
				if err != nil {
					logger.Error(ctx, "invalid URL", zap.Error(err))
					return err
				}

				// Marshal hashRequest to json
				reqJson, err := json.Marshal(hashRequest)
				if err != nil {
					logger.Error(ctx, "failed to unmarshall json string into hashRequest struct", zap.Error(err))
					return err
				}

				//Get pDNA and MD5 Hash
				req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v1/hash", bytes.NewBuffer(reqJson))
				if err != nil{
					logger.Error(ctx, "Error in creating a request to hasher service", zap.Error(err))
					return err
				}

				resp, err := httpClient.Do(req)

				if err != nil{
					logger.Error(ctx, "failed getting a response from hasher microservice", zap.Error(err))
					return err
				}

				//Log a non 200 response from hasher and continue
				if resp.StatusCode != 200{
					logger.Error(ctx,"Non 200 response from hasher service",zap.Error(err))
					ackErr := ackMessage(ctx,msg)
					if ackErr != nil{
						return ackErr
					}
					continue
				}
				body, err := ioutil.ReadAll(resp.Body)
				var hashedData types.HashResponse
				err = json.Unmarshal(body, &hashedData)

				objFingerprintRequest := types.FingerprintRequest{
					Path : hashedData.URL,
					MD5: hashedData.MD5,
					PhotoDNA: hashedData.PDNA,
					Product: scanRequestData.Product,
					Identifiers: types.AccountIdentifiers{
						ShopperId: scanRequestData.Identifiers.ShopperId,
						ContainerId: scanRequestData.Identifiers.ContainerId,
						Domain: scanRequestData.Identifiers.Domain,
						GUID: scanRequestData.Identifiers.GUID,
						XID: scanRequestData.Identifiers.XID,
					},

				}

				err = objFingerprintRequest.ValidateRequiredFields()
				if err != nil {
					logger.Error(ctx, "failed validating the FingerprintRequest attributes", zap.Error(err))
					return err
				}

				objProducer := Producer{
					env: c.env,
					uri: c.uri,
				}

				//Publish the new request to ThornWorker queue
				err = objProducer.Publish(&objFingerprintRequest)
				if err != nil {
					logger.Error(ctx, "failed publishing to the thornworker queue", zap.Error(err))
					return err
				}
			}

			err = ackMessage(ctx,msg)
			if err!=nil{
				return err
			}
		}
	}
}
