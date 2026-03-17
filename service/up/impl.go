package up

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/textproto"

	"github.com/bytedance/sonic"

	"github.com/volcengine/hiagent-go-sdk/types"
	"github.com/volcengine/hiagent-go-sdk/utils"
)

// UploadRaw ...
func (c Client) UploadRaw(ctx context.Context, req UploadRawRequest, reader io.Reader) (resp UploadRawResponse, err error) {
	var u = fmt.Sprintf("%s/up?Action=UploadRaw&Id=%s&Version=%s",
		c.host, req.ID, types.APIVersion_UP)
	if req.Expire != "" {
		u = fmt.Sprintf("%s/up?Action=UploadRaw&Id=%s&Version=%s&Expire=%s",
			c.host, req.ID, types.APIVersion_UP, req.Expire)
	}
	if req.Filename != nil {
		u = fmt.Sprintf("%s&Filename=%s", u, *req.Filename)
	}
	var header = make(http.Header)
	header.Set("X-Content-Sha256", req.Sha256)
	if req.ContentType == "" {
		header.Set(textproto.CanonicalMIMEHeaderKey("Content-Type"), "application/octet-stream")
	} else {
		header.Set(textproto.CanonicalMIMEHeaderKey("Content-Type"), req.ContentType)
	}
	var respBody io.Reader
	var code int
	if respBody, code, err = c.requester.MakeRequestRaw(ctx, u, http.MethodPost, reader, header); err != nil {
		return
	}
	var data []byte
	if data, err = io.ReadAll(respBody); err != nil {
		return
	}
	var resultCommonResponse types.CommonResponse
	if err = sonic.Unmarshal(data, &resultCommonResponse); err != nil {
		return
	}
	if code != 200 {
		if resultCommonResponse.ResponseMetadata.Error == nil {
			resultCommonResponse.ResponseMetadata.Error = &types.ErrorObj{CodeN: code}
		}
		err = utils.HandleError(resultCommonResponse)
		return
	}
	if err = sonic.Unmarshal(resultCommonResponse.Result, &resp); err != nil {
		return
	}
	return
}

// Delete ...
func (c Client) Delete(ctx context.Context, req DeleteRequest) (err error) {
	var u = fmt.Sprintf("%s/up?Action=Delete&Version=%s", c.host, types.APIVersion_UP)
	_, commResp, code, err := c.requester.MakeRequestWithContext(ctx, u, http.MethodPost, req)
	if err != nil {
		return
	}
	if code != 200 {
		err = utils.HandleError(commResp)
		return
	}
	return
}

// Download ...
func (c Client) Download(ctx context.Context, path, key string) (body io.Reader, err error) {
	var u = fmt.Sprintf("%s/down?Action=Download&Version=%s&Path=%s&Key=%s", c.host, types.APIVersion_UP, path, key)
	var code int
	if body, code, err = c.requester.MakeRequestRaw(ctx, u, http.MethodGet, nil); err != nil {
		return
	}
	if code != 200 {
		var data []byte
		if data, err = io.ReadAll(body); err != nil {
			return
		}
		var resultCommonResponse types.CommonResponse
		if err = sonic.Unmarshal(data, &resultCommonResponse); err != nil {
			return
		}
		if resultCommonResponse.ResponseMetadata.Error == nil {
			resultCommonResponse.ResponseMetadata.Error = &types.ErrorObj{CodeN: code}
		}
		err = utils.HandleError(resultCommonResponse)
		return
	}
	return
}

// DownloadKey ...
func (c Client) DownloadKey(ctx context.Context, path string) (resp DownloadKeyResponse, err error) {
	var u = fmt.Sprintf("%s/up?Action=DownloadKey&Version=%s&Path=%s", c.host, types.APIVersion_UP, path)
	headers := make(http.Header)
	headers.Add("Connection", "close")
	data, commResp, code, err := c.requester.MakeRequestWithContext(ctx, u, http.MethodPost, nil, headers)
	if err != nil {
		return
	}
	if code != 200 {
		err = utils.HandleError(commResp)
		return
	}
	if err = sonic.Unmarshal(data, &resp); err != nil {
		return
	}
	return
}
