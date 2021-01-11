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
	IMAGEEXCHANGENAME string      = "pdna-processor"
	VIDEOEXCHANGE     string      = "video-processor"
	MISCEXCHANGE      string      = "misc-processor"
	IMAGE_CONTENT     ContentType = "image"
	VIDEO_CONTENT     ContentType = "video"
	MISC_CONTENT      ContentType = "miscellaneous"
	VIDEO_HASHER_URL  string      = "http://localhost:8080/v1/hash/video"
	IMAGE_HASHER_URL  string      = "http://localhost:8080/v1/hash/image"
)

//getHashes accepts the url as input, calls the hasher service and
//returns the response as a byte sequence
func getHashes(ctx context.Context, url string, contentType ContentType) ([]byte, error) {
	var hasherURL string
	if contentType == VIDEO_CONTENT {
		hasherURL = VIDEO_HASHER_URL
	} else if contentType == IMAGE_CONTENT {
		hasherURL = IMAGE_HASHER_URL
	} else {
		return nil, errors.New("Unsupported file type by hasher microservice")
	}
	hashRequest := types.HashRequest{
		URL: url,
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
	if err != nil {
		logger.Error(ctx, "Error in obtaining hasher request json")
		return nil, err
	}

	//Get hashses from hashser micro service
	req, err := http.NewRequest(http.MethodPost, hasherURL, bytes.NewBuffer(reqJson))
	if err != nil {
		logger.Error(ctx, "Error in creating a request to hasher service", zap.Error(err))
		return nil, err
	}
	var httpClient = &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error(ctx, "failed getting a response from hasher microservice", zap.Error(err))
		return nil, err
	}

	//Log a non 200 response from hasher and continue
	if resp.StatusCode != 200 {
		logger.Error(ctx, "Non 200 response from hasher service", zap.Error(err))
		return nil, errors.New("Non 200 response")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(ctx, "Unable to convert response to byte sequence", zap.Error(err))
		return nil, err
	}
	return body, nil
}

// getContentType accepts an url as input, performs a get request and detects
// the content type based on the value of Content-Type header.
func getContentType(ctx context.Context, Url string, method string) (ContentType, error) {
	var httpClient = &http.Client{}
	req, err := http.NewRequest(method, Url, nil)
	if err != nil {
		logger.Error(ctx, "Error in get request", zap.Error(err))
		return "", err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get a response from %s", Url), zap.Error(err))
		return "", err
	}
	//Handling error response codes
	if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		logger.Error(ctx, fmt.Sprintf("Obtained status code %d", resp.StatusCode))
		return "", errors.New("Error status code")
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		logger.Info(ctx, fmt.Sprintf("Did not obtain content type for %s request", method))
		return "", nil
	}
	if strings.Contains(contentType, "image") {
		return IMAGE_CONTENT, nil
	} else if strings.Contains(contentType, "video") {
		return VIDEO_CONTENT, nil
	} else {
		return MISC_CONTENT, nil
	}
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
}

//ackMessage acknowledges the given amqp message
func (w Worker) ackMessage(msg amqp.Delivery) {
	if objErr := msg.Ack(false); objErr != nil {
		//A failure to ack, we send the cancel signal requesting all go routines to stop
		logger.Error(w.ctx, "error acknowledging message", zap.Error(objErr))
		w.cancelFunc()
	}
}

//rejectMessage rejects the given amqp message
func (w Worker) rejectMessage(msg amqp.Delivery) {
	if objErr := msg.Reject(false); objErr != nil {
		//A failure to nack, we send the cancel signal requesting all go routines to stop
		logger.Error(w.ctx, "error nacking message", zap.Error(objErr))
		w.cancelFunc()
	}
	logger.Info(w.ctx, "Rejected message")
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
			w.rejectMessage(imageMsg)
			continue
		}
		hasherResponse, err := getHashes(w.ctx, scanRequestData.URL, IMAGE_CONTENT)
		if err != nil {
			logger.Error(w.ctx, fmt.Sprintf("Hasher request failed for %s", scanRequestData.URL), zap.Error(err))
			w.rejectMessage(imageMsg)
			continue
		}
		var hashedData types.ImageHashResponse
		err = json.Unmarshal(hasherResponse, &hashedData)
		if err != nil {
			logger.Error(w.ctx, fmt.Sprintf("Failed to unmarshal hashser response for %s", scanRequestData.URL), zap.Error(err))
			w.rejectMessage(imageMsg)
			continue
		}
		imageFingerprintRequest := types.ImageFingerprintRequest{
			Path:        hashedData.URL,
			MD5:         hashedData.MD5,
			SHA1:        hashedData.SHA1,
			PhotoDNA:    hashedData.PDNA,
			Product:     scanRequestData.Product,
			Source:      "scan",
			Identifiers: scanRequestData.Identifiers,
		}
		err = imageFingerprintRequest.ValidateRequiredFields()
		if err != nil {
			logger.Error(w.ctx, "failed validating the FingerprintRequest attributes", zap.Error(err))
			w.rejectMessage(imageMsg)
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
			w.rejectMessage(imageMsg)
			w.cancelFunc()
			continue
		}

		w.ackMessage(imageMsg)
		logger.Info(w.ctx, fmt.Sprintf("Successfully processed %s image", scanRequestData.URL))
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
			w.rejectMessage(videoMsg)
			continue
		}
		hasherResponse, err := getHashes(w.ctx, scanRequestData.URL, VIDEO_CONTENT)
		if err != nil {
			logger.Error(w.ctx, "Hashser request failed", zap.Error(err))
			w.rejectMessage(videoMsg)
			continue
		}
		var videoHashedData types.VideoHashResponse
		err = json.Unmarshal(hasherResponse, &videoHashedData)
		if err != nil {
			logger.Error(w.ctx, "Failed to unmarshal hashser response")
			w.cancelFunc()
			continue
		}
		videoFingerprintRequest := types.VideoFingerprintRequest{
			Path:        videoHashedData.URL,
			MD5:         videoHashedData.MD5,
			SHA1:        videoHashedData.SHA1,
			Product:     scanRequestData.Product,
			Source:      "scan",
			Identifiers: scanRequestData.Identifiers,
		}

		//Publish the new request to hasher video queue
		json, err := json.Marshal(videoFingerprintRequest)
		if err != nil {
			log.Printf("unable to marshal message %s", err)
			w.cancelFunc()
			continue
		}
		logger.Debug(w.ctx, fmt.Sprintf("Producer json %s", string(json)))
		err = objProducer.Publish(w.ctx, json, VIDEOEXCHANGE)
		if err != nil {
			logger.Error(w.ctx, "failed publishing to video exchange", zap.Error(err))
			w.rejectMessage(videoMsg)
			continue
		}
		w.ackMessage(videoMsg)
		logger.Info(w.ctx, fmt.Sprintf("Successfully processed %s video", scanRequestData.URL))
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
			w.cancelFunc()
			continue
		}
		w.ackMessage(miscMsg)
		logger.Info(w.ctx, fmt.Sprintf("Successfully processed %s misc content", scanRequestData.URL))
		continue
	}
}

//contentTypeWorker listens to the job chan, detects the content type and routes the messages to imageIngestChan, videoIngestChan or miscIngestChan
func (w Worker) contentTypeWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Info(w.ctx, "Content type worker started")
	for msg := range w.jobsChan {
		scanRequestData := types.ScanRequest{}
		err := json.Unmarshal(msg.Body, &scanRequestData)
		if err != nil {
			logger.Error(w.ctx, "failed to unmarshall json string into scanRequestData struct", zap.Error(err))
			w.rejectMessage(msg)
			continue
		}
		logger.Debug(w.ctx, fmt.Sprintf("Scan URL: %s", scanRequestData.URL))
		contentType, err := getContentType(w.ctx, scanRequestData.URL, http.MethodHead)
		if err != nil || contentType == "" {
			logger.Info(w.ctx, fmt.Sprintf("Failed head request for %s url, reverting to get", scanRequestData.URL))
			contentType, err = getContentType(w.ctx, scanRequestData.URL, http.MethodGet)
		}
		if err != nil || contentType == "" {
			//Both head and get failed
			//Log error and ack message
			logger.Error(w.ctx, "Unable to detect content type", zap.Error(err))
			w.ackMessage(msg)
			continue
		}
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
