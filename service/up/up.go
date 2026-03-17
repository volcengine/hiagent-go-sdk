package up

import (
	"context"
	"io"

	"github.com/volcengine/hiagent-go-sdk/types"
	"github.com/volcengine/hiagent-go-sdk/utils"

	"github.com/volcengine/volc-sdk-golang/base"
)

// Client ...
type Client struct {
	host      string
	requester *utils.Requester
}

// New ...
func New(host, ak, sk string) Provider {
	requester := utils.New(base.Credentials{
		AccessKeyID: ak, SecretAccessKey: sk,
		Region:  types.DefaultRegion,
		Service: string(types.Service_UP),
	})
	cli := Client{requester: requester, host: host}
	return cli
}

// Provider up provider
type Provider interface {
	// UploadRaw ...
	UploadRaw(ctx context.Context, req UploadRawRequest, reader io.Reader) (resp UploadRawResponse, err error)
	// Delete ...
	Delete(ctx context.Context, req DeleteRequest) (err error)
	// Download ...
	Download(ctx context.Context, path, key string) (body io.Reader, err error)
	// DownloadKey ...
	DownloadKey(ctx context.Context, path string) (resp DownloadKeyResponse, err error)
}
