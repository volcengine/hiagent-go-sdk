package request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/volcengine/hiagent-go-sdk/hibot/internal/response"
	"github.com/volcengine/volc-sdk-golang/base"
)

// Config configures the TOP request executor shared by all resource clients.
type Config struct {
	Endpoint    string
	AccessKey   string
	SecretKey   string
	WorkspaceID string
	Region      string
	HTTPClient  *http.Client
}

// Action describes a TOP Action request.
type Action struct {
	Service string
	Version string
	Action  string
	Body    any
	Stream  bool
}

// Client signs, sends, and decodes TOP Action requests.
type Client struct {
	cfg Config
}

func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) DoAction(ctx context.Context, req Action, out any) error {
	body, err := c.marshalActionBody(req.Body)
	if err != nil {
		return err
	}
	httpReq, err := c.newHTTPRequest(ctx, req, bytes.NewReader(body), "application/json", nil)
	if err != nil {
		return err
	}
	resp, err := c.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, out)
}

func (c *Client) DoRawAction(ctx context.Context, req Action, body io.Reader, contentType string, query map[string]string, out any) error {
	httpReq, err := c.newHTTPRequest(ctx, req, body, contentType, query)
	if err != nil {
		return err
	}
	resp, err := c.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, out)
}

func (c *Client) NewStreamRequest(ctx context.Context, req Action) (*http.Request, error) {
	body, err := c.marshalActionBody(req.Body)
	if err != nil {
		return nil, err
	}
	req.Stream = true
	return c.newHTTPRequest(ctx, req, bytes.NewReader(body), "application/json", nil)
}

func (c *Client) DoStream(ctx context.Context, req Action) (*http.Response, error) {
	httpReq, err := c.NewStreamRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	streamClient := *c.cfg.HTTPClient
	streamClient.Timeout = 0
	return streamClient.Do(httpReq)
}

func (c *Client) newHTTPRequest(ctx context.Context, req Action, body io.Reader, contentType string, query map[string]string) (*http.Request, error) {
	u, err := url.Parse(c.cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("hibot: parse endpoint: %w", err)
	}
	// TOP 网关上 up 服务挂在子路径 /up，根路径不接受 up 的 Action。
	// 其他服务（hibot-server / hibot-gateway / aigw）仍通过根路径分发。
	if req.Service == "up" {
		u.Path = strings.TrimRight(u.Path, "/") + "/up"
	}
	q := u.Query()
	q.Set("Action", req.Action)
	q.Set("Version", req.Version)
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("X-Top-Service", req.Service)
	if req.Stream {
		httpReq.Header.Set("Accept", "text/event-stream")
	}

	signer := &base.Credentials{
		AccessKeyID:     c.cfg.AccessKey,
		SecretAccessKey: c.cfg.SecretKey,
		Region:          c.cfg.Region,
		Service:         req.Service,
	}
	return signer.Sign(httpReq), nil
}

func (c *Client) marshalActionBody(v any) ([]byte, error) {
	body, err := toMap(v)
	if err != nil {
		return nil, err
	}
	injectWorkspace(body, c.cfg.WorkspaceID)
	return json.Marshal(body)
}

func decodeResponse(resp *http.Response, out any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return response.DecodeTOP(resp.StatusCode, body, out)
}

func toMap(v any) (map[string]any, error) {
	if v == nil {
		return map[string]any{}, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("hibot: encode request: %w", err)
	}
	var body map[string]any
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, fmt.Errorf("hibot: normalize request: %w", err)
	}
	return body, nil
}

func injectWorkspace(body map[string]any, workspaceID string) {
	if workspaceID == "" {
		return
	}
	if workspaceMissing(body["WorkspaceID"]) {
		body["WorkspaceID"] = workspaceID
	}
}

func workspaceMissing(v any) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return s == ""
	}
	return false
}
