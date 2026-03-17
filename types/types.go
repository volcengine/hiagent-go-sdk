package types

import (
	"encoding/json"
)

const (
	// DefaultRegion default region
	DefaultRegion = "cn-north-1"
)

// APIVersion api version type
type APIVersion string

const (
	// APIVersion_UP is the up api version.
	APIVersion_UP APIVersion = "2022-01-01"
)

// Service service name type
type Service string

const (
	// Service_UP is the up service.
	Service_UP Service = "up"
)

// IamConfig ...
type IamConfig struct {
	Ak      string
	Sk      string
	Region  string
	Service Service
}

// CommonResponse is a common response
type CommonResponse struct {
	ResponseMetadata ResponseMetadata
	Result           json.RawMessage `json:"Result,omitempty"`
}

// ResponseMetadata is a response metadata
type ResponseMetadata struct {
	RequestId string
	Action    string    `json:",omitempty"`
	Version   string    `json:",omitempty"`
	Service   string    `json:",omitempty"`
	Region    string    `json:",omitempty"`
	Error     *ErrorObj `json:",omitempty"`
}

// ErrorObj is an error object
type ErrorObj struct {
	CodeN   int    `json:",omitempty"`
	Code    string `json:",omitempty"`
	Message string `json:",omitempty"`
}
