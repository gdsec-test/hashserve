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

type ThornWorkerRequest struct {
	Path string `json:"path"`
	PhotoDNA string `json:"photoDNA"`
	MD5 string `json:"MD5"`
	accountIdentifiers AccountIdentifiers
}

func(tr *ThornWorkerRequest) ValidateRequiredFields() error {
	if (AccountIdentifiers{}) == tr.accountIdentifiers {
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