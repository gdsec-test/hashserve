package types

import (
	"errors"
)

type AccountIdentifiers struct {
	ShopperId string `json:"shopperID"`
	ContainerId string `json:"containerID"`
	Domain string `json:"domain"`
	GUID string `json:"GUID"`
	XID string `json:"XID"`
}

type FingerprintRequest struct {
	Path string `json:"path"`
	PhotoDNA string `json:"photoDNA"`
	MD5 string `json:"MD5"`
	Product string `json:"product"`
	Identifiers AccountIdentifiers `json:"accountIdentifiers"`
}

// ScanRequest represents the full request made by a product
// to submit potential CSAM
type ScanRequest struct {
	Identifiers AccountIdentifiers `json:"accountIdentifiers"`
	URL string `json:"url"`
}

func(tr *FingerprintRequest) ValidateRequiredFields() error {
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