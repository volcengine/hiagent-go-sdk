package utils

import (
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/volcengine/hiagent-go-sdk/types"
)

// Error ...
type Error struct {
	HTTPCode int32
	Code     string
	Message  string
	Data     map[string]string
}

// Error...
func (e *Error) Error() string {
	return e.String()
}

// String...
func (e *Error) String() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Error(%+v)", *e)
}

// HandleError ...
var HandleError = func(commResp types.CommonResponse) error {
	var m = map[string]string{}
	_ = sonic.Unmarshal(commResp.Result, &m)
	m["request_id"] = commResp.ResponseMetadata.RequestId
	return &Error{
		HTTPCode: int32(commResp.ResponseMetadata.Error.CodeN),
		Code:     commResp.ResponseMetadata.Error.Code,
		Message:  commResp.ResponseMetadata.Error.Message,
		Data:     m,
	}
}
