package hibot

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/volcengine/hiagent-go-sdk/hibot/internal/request"
	"github.com/volcengine/hiagent-go-sdk/hibot/internal/version"
	v1 "github.com/volcengine/hiagent-go-sdk/hibot/v1"
)

const (
	defaultRegion         = version.DefaultRegion
	defaultServerService  = version.ServerService
	defaultGatewayService = version.GatewayService
	defaultModelService   = version.AIGWService
	defaultUpService      = version.UPService
	defaultV1Version      = version.V1
	defaultModelVersion   = version.Model
	defaultServerVersion  = version.Server
	defaultChatVersion    = version.Chat
	defaultUpVersion      = version.UP
)

// Config configures the Hibot SDK client.
type Config struct {
	Endpoint    string
	AccessKey   string
	SecretKey   string
	WorkspaceID string
	Region      string

	HTTPClient *http.Client

	ServerService  string
	GatewayService string
	ModelService   string
	UpService      string
}

// NewClient creates a Hibot SDK client.
func NewClient(cfg Config) (*Client, error) {
	cfg.Endpoint = cleanString(cfg.Endpoint)
	cfg.AccessKey = cleanString(cfg.AccessKey)
	cfg.SecretKey = cleanString(cfg.SecretKey)
	cfg.WorkspaceID = cleanString(cfg.WorkspaceID)
	cfg.Region = cleanString(cfg.Region)
	cfg.ServerService = cleanString(cfg.ServerService)
	cfg.GatewayService = cleanString(cfg.GatewayService)
	cfg.ModelService = cleanString(cfg.ModelService)
	cfg.UpService = cleanString(cfg.UpService)

	if cfg.Endpoint == "" {
		return nil, errors.New("hibot: endpoint is required")
	}
	if cfg.AccessKey == "" {
		return nil, errors.New("hibot: access key is required")
	}
	if cfg.SecretKey == "" {
		return nil, errors.New("hibot: secret key is required")
	}
	if cfg.WorkspaceID == "" {
		return nil, errors.New("hibot: workspace id is required")
	}
	if cfg.Region == "" {
		cfg.Region = defaultRegion
	}
	if cfg.ServerService == "" {
		cfg.ServerService = defaultServerService
	}
	if cfg.GatewayService == "" {
		cfg.GatewayService = defaultGatewayService
	}
	if cfg.ModelService == "" {
		cfg.ModelService = defaultModelService
	}
	if cfg.UpService == "" {
		cfg.UpService = defaultUpService
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	requester := request.NewClient(request.Config{
		Endpoint:    cfg.Endpoint,
		AccessKey:   cfg.AccessKey,
		SecretKey:   cfg.SecretKey,
		WorkspaceID: cfg.WorkspaceID,
		Region:      cfg.Region,
		HTTPClient:  cfg.HTTPClient,
	})
	c := &Client{}
	c.V1 = v1.NewClient(requester, v1.Services{
		Server:  cfg.ServerService,
		Gateway: cfg.GatewayService,
		Model:   cfg.ModelService,
		UP:      cfg.UpService,
	})
	return c, nil
}

// String returns a pointer to v.
func String(v string) *string {
	return &v
}

func cleanString(v string) string {
	return strings.TrimSpace(v)
}
