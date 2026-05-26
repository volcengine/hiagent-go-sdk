package hibot

import (
	"github.com/volcengine/hiagent-go-sdk/hibot/v1"
)

// Client is the root Hibot SDK client.
type Client struct {
	V1 *v1.Client
}
