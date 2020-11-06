package types

import (
	"errors"
	"net/url"
)

// AccountIdentifiers structure
type AccountIdentifiers struct {
	ShopperId string `json:"shopperID"`
	ContainerId string `json:"containerID"`
	Domain string `json:"domain"`
	GUID string `json:"GUID"`
	XID string `json:"XID"`
}

// ImageFingerprintRequest structure
type ImageFingerprintRequest struct {
	Path string `json:"path"`
	PhotoDNA string `json:"photoDNA"`
	MD5 string `json:"MD5"`
	SHA1 string `json:"SHA1"`
	Product string `json:"product"`
	Identifiers AccountIdentifiers `json:"accountIdentifiers"`
}

//VideoFingerPrintRequest structure
type VideoFingerprintRequest struct{
	Path string `json:"path"`
	MD5 string `json:"MD5"`
	SHA1 string `json:"SHA1"`
	Product string `json:"product"`
	Identifiers AccountIdentifiers `json:"accountIdentifiers"`
}

// ScanRequest represents the full request made by a product
// to submit potential CSAM
type ScanRequest struct {
	Identifiers AccountIdentifiers `json:"accountIdentifiers"`
	URL string `json:"url"`
	Product string `json:"product"`
}

// HashRequest represents the full request made by hashserve to Hasher microservice
type HashRequest struct {
	URL string `json:"URL"`
}

// ImageHashResponse represents the full response received from Hasher microservice
type ImageHashResponse struct {
	URL string `json:"URL"`
	PDNA string `json:"PDNA"`
	MD5 string `json:"MD5"`
	SHA1 string `json:"SHA1"`
}

// VideoHashResponse represents the full response received from Hasher microservice
type VideoHashResponse struct {
	URL string `json:"URL"`
	MD5 string `json:"MD5"`
	SHA1 string `json:"SHA1"`
}

// function to validate the URL being sent  over to hasher microservice
func(hr *HashRequest) ValidateRequiredFields() error {
	//ParseRequestURI parses rawurl into a URL structure. It assumes that rawurl was received in an HTTP request,
	//so the rawurl is interpreted only as an absolute URI or an absolute path.
	//The string rawurl is assumed not to have a #fragment suffix.
	_, err := url.ParseRequestURI(hr.URL)
	if err != nil {
		return errors.New("invalid URL")
	}
	return nil
}

// function to validate the fields before publishing the message to the thornworker queue.
func(tr *ImageFingerprintRequest) ValidateRequiredFields() error {
	if (AccountIdentifiers{}) == tr.Identifiers {
		return errors.New("missing account identifiers")
	}

	if tr.Path == "" {
		return errors.New("missing path")
	}

	if tr.PhotoDNA == "" && tr.MD5 == "" {
		return errors.New("missing photoDNA and MD5")
	}

	return nil
}