package rabbitmq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/streadway/amqp"
	"github.secureserver.net/digital-crimes/hashserve/pkg/logger"
	"github.secureserver.net/digital-crimes/hashserve/pkg/types"
	"go.uber.org/zap"
)

type ContentType string

const (
	IMAGEEXCHANGENAME                   string      = "pdna-processor"
	VIDEOEXCHANGE                       string      = "video-processor"
	MISCEXCHANGE                        string      = "misc-processor"
	RETRYEXCHANGE                       string      = "hashserve-dlq"
	IMAGE_CONTENT                       ContentType = "image"
	VIDEO_CONTENT                       ContentType = "video"
	MISC_CONTENT                        ContentType = "miscellaneous"
	VIDEO_HASHER_URL                    string      = "http://localhost:8080/v1/hash/video"
	IMAGE_HASHER_URL                    string      = "http://localhost:8080/v1/hash/image"
	DOWNLOAD_FAILED_FILE_NOT_FOUND_CODE int         = 4
	HASH_SUCCESS_STATUS_CODE            int         = 1
)

//getHashes accepts the url as input, calls the hasher service and
//returns the response as a byte sequence
func getHashes(ctx context.Context, url string, cert string, contentType ContentType) ([]byte, error) {
	var hasherURL string
	if contentType == VIDEO_CONTENT {
		hasherURL = VIDEO_HASHER_URL
	} else if contentType == IMAGE_CONTENT {
		hasherURL = IMAGE_HASHER_URL
	} else {
		return nil, errors.New("Unsupported file type by hasher microservice")
	}
	hashRequest := types.HashRequest{
		URL:  url,
		Cert: cert,
	}
	err := hashRequest.ValidateRequiredFields()
	if err != nil {
		logger.Error(ctx, "invalid URL", zap.Error(err))
		return nil, err
	}

	// Marshal hashRequest to json
	reqJson, err := json.Marshal(hashRequest)
	if err != nil {
		logger.Error(ctx, "failed to unmarshall json string into hashRequest struct", zap.Error(err))
		return nil, err
	}

	//Get hashses from hashser micro service
	req, err := http.NewRequest(http.MethodPost, hasherURL, bytes.NewBuffer(reqJson))
	if err != nil {
		logger.Error(ctx, "Error in creating a request to hasher service", zap.Error(err))
		return nil, err
	}
	var httpClient = &http.Client{
		Timeout: 2 * time.Minute,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed getting a response from hasher microservice. Request JSON: %s", string(reqJson)), zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	logger.Debug(ctx, fmt.Sprintf("Hasher status code: %d, Body: %s", resp.StatusCode, string(body)))
	if err != nil {
		logger.Error(ctx, "Unable to convert response to byte sequence", zap.Error(err))
		return nil, err
	}
	return body, nil
}

// getContentType checks if the file extension in url matches the miscellaneous or video extension
// list. If no matches are found, its assumed that the given content is of image type
func getContentType(ctx context.Context, Url string) ContentType {
	miscContentExtension := []string{".pdf", ".svg", ".doc", ".docx"}
	videoContentExtension := []string{".mp4", ".wav"}
	for _, content := range miscContentExtension {
		if strings.HasSuffix(Url, content) {
			return MISC_CONTENT
		}
	}
	for _, content := range videoContentExtension {
		if strings.HasSuffix(Url, content) {
			return VIDEO_CONTENT
		}
	}
	return IMAGE_CONTENT
}

/*Worker is a wrapper around the different worker go routines.
amqp messages are fed to the jobsChan where the content type is detected
and routed appropriately to imageIngestChan, videoIngestChan or miscIngestChan.*/
type Worker struct {
	imageIngestChan chan amqp.Delivery
	videoIngestChan chan amqp.Delivery
	miscIngestChan  chan amqp.Delivery
	jobsChan        chan amqp.Delivery
	ctx             context.Context
	cancelFunc      context.CancelFunc
	env             string
	uri             string
	conn            *Connection
	maxRetryCount   int
}

//ackMessage acknowledges the given amqp message
func (w Worker) ackMessage(msg amqp.Delivery) {
	if objErr := msg.Ack(false); objErr != nil {
		//A failure to ack, we send the cancel signal requesting all go routines to stop
		logger.Error(w.ctx, "error acknowledging message", zap.Error(objErr))
		w.cancelFunc()
	}
}

//rejectMessageWithoutRequeue rejects the given amqp message
func (w Worker) rejectMessageWithoutRequeue(msg amqp.Delivery) {
	if objErr := msg.Reject(false); objErr != nil {
		//A failure to nack, we send the cancel signal requesting all go routines to stop
		logger.Error(w.ctx, "error nacking message", zap.Error(objErr))
		w.cancelFunc()
	}
}

/*imageWorkerFunc listens to imageIngestChan, calls the hasher microservice to get hashes
and routes response to image exchange.*/
func (w Worker) imageWorkerFunc(wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Info(w.ctx, "Image worker started")
	objProducer, err := NewProducer(w.ctx, w.env, w.conn)
	if err != nil {
		logger.Error(w.ctx, "Unable to create a producer", zap.Error(err))
		w.cancelFunc()
		return
	}
	defer objProducer.ch.Close()
	for imageMsg := range w.imageIngestChan {
		logger.Debug(w.ctx, "Image channel started")
		scanRequestData := types.ScanRequest{}
		err := json.Unmarshal(imageMsg.Body, &scanRequestData)
		//If unable to unmarshal the message into scanRequestData, log the error.
		if err != nil {
			logger.Error(w.ctx, "failed to unmarshall json string into scanRequestData struct", zap.Error(err))
			w.rejectMessageWithoutRequeue(imageMsg)
			continue
		}
		hasherResponse, err := getHashes(w.ctx, scanRequestData.URL, scanRequestData.Cert, IMAGE_CONTENT)
		var hashedData types.ImageHashResponse
		errUnmarshal := json.Unmarshal(hasherResponse, &hashedData)
		// We shouldn't encounter this error ideally
		if errUnmarshal != nil {
			logger.Error(w.ctx, "Failed to unmarshal JSON", zap.Error(err))
			w.ackMessage(imageMsg)
			continue
		}
		// This section of the code deals with retrying using a dead letter queue.
		// Reject message if hasher either returns a file not found error or if retry count is greater than or equal to max retry count.
		// Reque if the retry count is below max retry count and hasher returns a status code other than file not found or hash success.
		if err == nil && hashedData.StatusCode == DOWNLOAD_FAILED_FILE_NOT_FOUND_CODE {
			logger.Error(w.ctx, fmt.Sprintf("Obtained file not found status code for %s. Rejecting message", scanRequestData.URL))
			w.ackMessage(imageMsg)
			continue
		} else if hashedData.StatusCode != HASH_SUCCESS_STATUS_CODE && scanRequestData.RetryCount >= w.maxRetryCount {
			logger.Error(w.ctx, fmt.Sprintf("Max retry count reached for %s. Rejecting message", scanRequestData.URL))
			w.ackMessage(imageMsg)
			continue
		} else if hashedData.StatusCode != HASH_SUCCESS_STATUS_CODE || err != nil {
			// Requeue in dead letter queue
			scanRequestData.RetryCount = scanRequestData.RetryCount + 1
			scanRequestData.PublishTime = time.Now().Format(time.RFC3339)
			json, _ := json.Marshal(scanRequestData)
			err = objProducer.Publish(w.ctx, json, RETRYEXCHANGE)
			if err != nil {
				logger.Error(w.ctx, "failed publishing to the retry queue", zap.Error(err))
				w.cancelFunc()
				continue
			}
			logger.Error(w.ctx, fmt.Sprintf("Obtained status message: %s.%s URL published for retry", hashedData.StatusMessage, scanRequestData.URL))
			w.ackMessage(imageMsg)
			continue
		}
		imageFingerprintRequest := types.ImageFingerprintRequest{
			Path:        hashedData.URL,
			MD5:         hashedData.Hashes.MD5,
			SHA1:        hashedData.Hashes.SHA1,
			PhotoDNA:    hashedData.Hashes.PDNA,
			Product:     scanRequestData.Product,
			Source:      "scan",
			Identifiers: scanRequestData.Identifiers,
		}
		err = imageFingerprintRequest.ValidateRequiredFields()
		if err != nil {
			logger.Error(w.ctx, "failed validating the FingerprintRequest attributes", zap.Error(err))
			w.rejectMessageWithoutRequeue(imageMsg)
			continue
		}

		//Publish the new request to ThornWorker queue
		json, err := json.Marshal(imageFingerprintRequest)
		logger.Debug(w.ctx, fmt.Sprintf("Producer json %s", string(json)))
		if err != nil {
			log.Printf("unable to marshal message %s", err)
			w.cancelFunc()
			continue
		}
		err = objProducer.Publish(w.ctx, json, IMAGEEXCHANGENAME)
		if err != nil {
			logger.Error(w.ctx, "failed publishing to the thornworker queue", zap.Error(err))
			w.cancelFunc()
			continue
		}

		w.ackMessage(imageMsg)
		logger.Debug(w.ctx, fmt.Sprintf("Successfully processed %s image", scanRequestData.URL))
	}
}

/*videoWorkerFunc listens to videoIngestChan, calls the hasher microservice to get hashes
and routes response to video exchange.*/
func (w Worker) videoWorkerFunc(wg *sync.WaitGroup) {
	defer wg.Done()
	objProducer, err := NewProducer(w.ctx, w.env, w.conn)
	if err != nil {
		logger.Error(w.ctx, "Unable to create a producer", zap.Error(err))
		w.cancelFunc()
		return
	}
	defer objProducer.ch.Close()
	logger.Info(w.ctx, "Video worker started")
	for videoMsg := range w.videoIngestChan {
		logger.Debug(w.ctx, "Video channel started")
		scanRequestData := types.ScanRequest{}
		err := json.Unmarshal(videoMsg.Body, &scanRequestData)
		//If unable to unmarshal the message into scanRequestData, log the error.
		if err != nil {
			logger.Error(w.ctx, "failed to unmarshall json string into scanRequestData struct", zap.Error(err))
			w.rejectMessageWithoutRequeue(videoMsg)
			continue
		}
		w.ackMessage(videoMsg)
		logger.Debug(w.ctx, fmt.Sprintf("Successfully processed %s video", scanRequestData.URL))
	}
}

//miscWorkerFunc listens to miscIngestChan
func (w Worker) miscWorkerFunc(wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Info(w.ctx, "Misc worker started")
	for miscMsg := range w.miscIngestChan {
		logger.Debug(w.ctx, "Miscellaneous channel started")
		scanRequestData := types.ScanRequest{}
		err := json.Unmarshal(miscMsg.Body, &scanRequestData)
		if err != nil {
			log.Printf("unable to marshal message %s", err)
			w.rejectMessageWithoutRequeue(miscMsg)
			continue
		}
		w.ackMessage(miscMsg)
		logger.Debug(w.ctx, fmt.Sprintf("Successfully processed %s misc content", scanRequestData.URL))
		continue
	}
}

//contentTypeWorker listens to the job chan, detects the content type and routes the messages to imageIngestChan, videoIngestChan or miscIngestChan
func (w Worker) contentTypeWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Info(w.ctx, "Content type worker started*")
	for msg := range w.jobsChan {
		scanRequestData := types.ScanRequest{}
		err := json.Unmarshal(msg.Body, &scanRequestData)
		if err != nil {
			logger.Error(w.ctx, "failed to unmarshall json string into scanRequestData struct", zap.Error(err))
			w.rejectMessageWithoutRequeue(msg)
			continue
		}
		contentType := getContentType(w.ctx, scanRequestData.URL)
		logger.Debug(w.ctx, fmt.Sprintf("Scan URL: %s, Content type: %s", scanRequestData.URL, contentType))
		if contentType == IMAGE_CONTENT {
			logger.Debug(w.ctx, "Image content detected")
			w.imageIngestChan <- msg
		} else if contentType == VIDEO_CONTENT {
			logger.Debug(w.ctx, "Video content detected")
			w.videoIngestChan <- msg
		} else if contentType == MISC_CONTENT {
			logger.Debug(w.ctx, "Misc content detected")
			w.miscIngestChan <- msg
		}
	}
}
