package v1

import (
	"context"
	"errors"
	"io"

	"github.com/volcengine/hiagent-go-sdk/hibot/internal/request"
	"github.com/volcengine/hiagent-go-sdk/hibot/internal/version"
)

type UploadsService struct{ client *Client }

func (s *UploadsService) UploadBlob(ctx context.Context, params V1UploadBlobParams, body io.Reader) (*V1UploadBlob, error) {
	if params.Filename == "" {
		return nil, errors.New("hibot: upload filename is required")
	}
	var result V1UploadBlob
	if err := s.client.requester.DoRawAction(ctx, request.Action{
		Service: s.client.services.UP,
		Version: version.UP,
		Action:  "UploadBlob",
	}, body, params.ContentType, map[string]string{
		"Filename": params.Filename,
	}, &result); err != nil {
		return nil, err
	}
	if result.BlobID == "" {
		return nil, errors.New("hibot: upload blob response missing BlobID")
	}
	return &result, nil
}
