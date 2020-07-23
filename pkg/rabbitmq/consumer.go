package rabbitmq

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.secureserver.net/digital-crimes/hashserve/pkg/logger"
	"github.secureserver.net/digital-crimes/hashserve/pkg/types"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
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

// Serve creates a new Connection and opens a new Channel to a RabbitMQ Broker.
func (c *Consumer) Serve(ctx context.Context) error {

	conn, err := Dial(c.uri)

	logger.Info(ctx, "URI",zap.String("URI", c.uri))
	logger.Info(ctx, "ENV",zap.String("URI", c.env))

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

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-deliveries:
			/*  Basic Functionalities of the following goRoutine
				1) downloadFile - Downloads the content provided by the url in /tmp/pdna/ folder
				2) generateMD5 - Generates an MD5 Hash of the downloaded file
				3) generatePhotoDNA - Generates the photoDNA Hash of the downloaded file.
				4) deleteFile - Deletes the downloaded file
				5) creates and validates the ThornWorkerRequest
				6) Publish the new request to ThornWorker queue
			*/

			message := string(msg.Body)
			logger.Info(ctx, "Msg received",zap.String("msg", message))

			if objErr := msg.Ack(false); objErr != nil {
				logger.Error(ctx,"Error acknowledging message", zap.Error(objErr))
				return objErr
			}


			scanRequestData := types.ScanRequest{}
			err := json.Unmarshal([]byte(message), &scanRequestData)

			//If unable to unmarshal the message into scanRequestData, log the error and acknowledge the message.
			if err != nil {
				logger.Error(ctx, "failed to unmarshall json string into scanRequestData struct", zap.Error(err))
			}else{

				//Generates the name of the file : sha of the URL would be used as the filename to avoid overwriting
				fileName := buildFileName(scanRequestData.URL)
				path := os.Getenv("DOWNLOAD_FILE_LOC") + "/" + fileName

				//Downloads the content provided by the url in /app/tmp/pdna/ folder
				err = downloadFile(scanRequestData.URL, path)
				if err != nil {
					logger.Error(ctx, "failed to download the file from the url", zap.Error(err))
				}else{
					var objError error
					objError = nil

					//Generates an MD5 Hash of the downloaded file
					md5Hash, err := generateMD5Hash(path)
					if err != nil {
						logger.Error(ctx, "failed to generate MD5 hash of the file", zap.Error(err))
						objError = err
					}

					//Generates the photoDNA Hash of the downloaded file.
					//PhotoDNA binaries should be located in /app/pdna/bin/java path
					photoDNAHash, err := generatePhotoDNAHash(ctx, path)
					if err != nil {
						logger.Error(ctx, "failed to generate photoDNA hash of the file", zap.Error(err))
						objError = err
					}

					//create and validate the new ThornWorkerRequest
					objFingerprintRequest := types.FingerprintRequest{
						Path : scanRequestData.URL,
						MD5: md5Hash,
						PhotoDNA: photoDNAHash,
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
						objError = err
					}

					//Publish to ThornWorker queue only if no issues were encountered.
					if objError == nil{
						// Same Environment and URL as Consumer - but different exchange and Queue
						objProducer := Producer{
							env: c.env,
							uri: c.uri,
						}
						logger.Info(ctx, "Msg being send",zap.Any("msg: ", objFingerprintRequest))

						//Publish the new request to ThornWorker queue
						err = objProducer.Publish(&objFingerprintRequest)
						if err != nil {
							logger.Error(ctx, "failed publishing to the thornworker queue", zap.Error(err))
						}

						//Delete the file -if present
						err = deleteFile(path)
						if err != nil {
							logger.Error(ctx, "failed to delete the file: " + path, zap.Error(err))
						}
					}

				}
			}
		}
	}
}

func generatePhotoDNAHash(ctx context.Context, path string) (string, error) {
	cmd := exec.Command("java", "-cp", os.Getenv("CLASSPATH"), "GenerateHashes", path)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return "", err
	}

	pDNAOutput := string(out)
	pDNAHash := strings.Split(pDNAOutput, ",")
	if len(pDNAHash) != 145 {
		err := errors.New("invalid photoDNA")
		return "",err
	}
	photoDNAHash := strings.Join(pDNAHash[1:], ",")
	return photoDNAHash, nil
}

func deleteFile(path string) error {
	if fileExists(path) {
		var err = os.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func generateMD5Hash(fileName string) (string, error) {

	if fileExists(fileName) {
		// Reading From File
		file, err := os.Open(fileName)
		if err != nil {
			return "", err
		}

		hash := md5.New()
		//Copy the file in the hash interface and check for any error
		if _, err := io.Copy(hash, file); err != nil {
			return "", err
		}

		//Get the 16 bytes hash
		hashInBytesFile := hash.Sum(nil)[:16]

		//Convert the bytes to a string
		returnMD5StringFile := hex.EncodeToString(hashInBytesFile)
		file.Close()

		return returnMD5StringFile, nil
	}
	err := errors.New(fileName + " does not exist")
	return "", err
}

func buildFileName(fullURLFile string) string {

	extension:= fullURLFile[strings.LastIndex(fullURLFile, "."):]

	h := sha1.New()
	h.Write([]byte(fullURLFile))
	var fileName = hex.EncodeToString(h.Sum(nil)) + extension

	return fileName
}

func downloadFile(url, fileName string) error{
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	//local file in the associated path is created
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}