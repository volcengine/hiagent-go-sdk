package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/textproto"
	"time"

	"github.com/bytedance/sonic"

	"github.com/volcengine/hiagent-go-sdk/types"
	"github.com/volcengine/volc-sdk-golang/base"
)

// Requester ...
type Requester struct {
	credentials base.Credentials
}

// New ...
func New(credentials base.Credentials) *Requester {
	return &Requester{
		credentials: credentials,
	}
}

// MakeRequest sign request with ak, sk and send request to iam host
// Note: public cloud send request to top host
func (r *Requester) MakeRequest(u string, method string, reqBody any, headers ...http.Header) (respBody []byte, resultCommonResponse types.CommonResponse, code int, err error) {
	return r.MakeRequestWithContext(context.Background(), u, method, reqBody, headers...)
}

// MakeRequestWithContext sign request with ak, sk and send request to iam host
// Note: public cloud send request to top host
func (r *Requester) MakeRequestWithContext(ctx context.Context, u string, method string, reqBody any, headers ...http.Header) (respBody []byte, resultCommonResponse types.CommonResponse, code int, err error) {
	var reqBodyBytes = []byte{}
	if reqBody != nil {
		if reqBodyBytes, err = sonic.Marshal(reqBody); err != nil {
			return
		}
	}

	var body io.Reader
	if body, code, err = r.MakeRequestRaw(ctx, u, method, bytes.NewBuffer(reqBodyBytes), headers...); err != nil {
		return
	}
	var rawRespBody []byte
	if rawRespBody, err = io.ReadAll(body); err != nil {
		return
	}

	if err = sonic.Unmarshal(rawRespBody, &resultCommonResponse); err != nil {
		return
	}
	if code != 200 {
		if resultCommonResponse.ResponseMetadata.Error == nil {
			resultCommonResponse.ResponseMetadata.Error = &types.ErrorObj{}
		}
		resultCommonResponse.ResponseMetadata.Error.CodeN = code
	}
	respBody = resultCommonResponse.Result
	return
}

// MakeRequestRaw sign request with ak, sk and send request to iam host
// Note: public cloud send request to top host
func (r *Requester) MakeRequestRaw(ctx context.Context, u string, method string, reqBody io.Reader, headers ...http.Header) (body io.Reader, code int, err error) {
	var req *http.Request
	if method == http.MethodGet {
		if req, err = http.NewRequestWithContext(ctx, method, u, nil); err != nil {
			return
		}
	} else {
		if req, err = http.NewRequestWithContext(ctx, method, u, reqBody); err != nil {
			return
		}
	}
	if req.Header.Get(textproto.CanonicalMIMEHeaderKey("Content-Type")) == "" {
		req.Header.Set(textproto.CanonicalMIMEHeaderKey("Content-Type"), "application/json")
	}
	if len(headers) == 1 {
		for k, v := range headers[0] {
			req.Header.Set(k, v[0])
		}
	}
	req = r.credentials.Sign(req)

	var transport = &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	var c = http.Client{Transport: transport}
	var resp *http.Response
	if resp, err = c.Do(req); err != nil {
		return
	}

	code = resp.StatusCode
	body = resp.Body
	return
}
