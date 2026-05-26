package response

import (
	"encoding/json"
	"fmt"
)

// APIError is returned for Hibot TOP API errors.
type APIError struct {
	StatusCode int
	RequestID  string
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Code == "" && e.Message == "" {
		return fmt.Sprintf("hibot: api error status=%d request_id=%s", e.StatusCode, e.RequestID)
	}
	return fmt.Sprintf("hibot: api error status=%d request_id=%s code=%s message=%s", e.StatusCode, e.RequestID, e.Code, e.Message)
}

type topEnvelope struct {
	ResponseMetadata struct {
		RequestID string       `json:"RequestId"`
		Error     *topAPIError `json:"Error"`
	} `json:"ResponseMetadata"`
	Result json.RawMessage `json:"Result"`
}

type topAPIError struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

// DecodeTOP decodes a TOP envelope and unwraps Result into out.
func DecodeTOP(statusCode int, body []byte, out any) error {
	var env topEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		if statusCode >= 400 {
			return &APIError{StatusCode: statusCode, Message: string(body)}
		}
		return fmt.Errorf("hibot: decode response: %w", err)
	}
	if statusCode >= 400 || env.ResponseMetadata.Error != nil {
		apiErr := &APIError{
			StatusCode: statusCode,
			RequestID:  env.ResponseMetadata.RequestID,
		}
		if env.ResponseMetadata.Error != nil {
			apiErr.Code = env.ResponseMetadata.Error.Code
			apiErr.Message = env.ResponseMetadata.Error.Message
		}
		return apiErr
	}
	if out == nil || len(env.Result) == 0 || string(env.Result) == "null" {
		return nil
	}
	if err := json.Unmarshal(env.Result, out); err != nil {
		return fmt.Errorf("hibot: decode result: %w", err)
	}
	return nil
}
