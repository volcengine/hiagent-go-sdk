package v1

import "encoding/json"

type V1MCP struct {
	ID                   string            `json:"ID"`
	Name                 string            `json:"Name,omitempty"`
	Description          string            `json:"Description,omitempty"`
	Transport            string            `json:"Transport,omitempty"`
	Endpoint             string            `json:"URL,omitempty"`
	Headers              map[string]string `json:"Headers,omitempty"`
	Env                  map[string]string `json:"Env,omitempty"`
	Command              string            `json:"Command,omitempty"`
	Args                 []string          `json:"Args,omitempty"`
	AuthType             string            `json:"AuthType,omitempty"`
	CredentialProviderID string            `json:"CredentialProviderID,omitempty"`
	ToolAllowlist        []string          `json:"ToolAllowlist,omitempty"`
	ToolDenylist         []string          `json:"ToolDenylist,omitempty"`
	ToolPrefix           string            `json:"ToolPrefix,omitempty"`
	Timeout              int64             `json:"Timeout,omitempty"`
	Status               string            `json:"Status,omitempty"`
	Source               string            `json:"Source,omitempty"`
	CreatedAt            string            `json:"CreatedAt,omitempty"`
	UpdatedAt            string            `json:"UpdatedAt,omitempty"`
}

type V1MCPNewParams struct {
	WorkspaceID      string                      `json:"WorkspaceID,omitempty"`
	Name             string                      `json:"Name,omitempty"`
	Description      string                      `json:"Description,omitempty"`
	Transport        string                      `json:"Transport,omitempty"`
	Endpoint         string                      `json:"URL,omitempty"`
	Headers          map[string]string           `json:"Headers,omitempty"`
	Env              map[string]string           `json:"Env,omitempty"`
	Command          string                      `json:"Command,omitempty"`
	Args             []string                    `json:"Args,omitempty"`
	AuthType         string                      `json:"AuthType,omitempty"`
	CredentialConfig *V1MCPCredentialInputParams `json:"CredentialConfig,omitempty"`
	ToolAllowlist    []string                    `json:"ToolAllowlist,omitempty"`
	ToolDenylist     []string                    `json:"ToolDenylist,omitempty"`
	ToolPrefix       string                      `json:"ToolPrefix,omitempty"`
	Timeout          int64                       `json:"Timeout,omitempty"`
	Source           string                      `json:"Source,omitempty"`
}

type V1MCPResolveParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	ID          string `json:"ID,omitempty"`
	Name        string `json:"Name,omitempty"`
}

type V1MCPListParams struct {
	WorkspaceID string       `json:"WorkspaceID,omitempty"`
	Keyword     string       `json:"Keyword,omitempty"`
	Status      string       `json:"Status,omitempty"`
	Source      string       `json:"Source,omitempty"`
	Page        *V1PageInput `json:"Page,omitempty"`
}

type V1MCPGetParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	ID          string `json:"ID,omitempty"`
}

type V1MCPUpdateParams struct {
	WorkspaceID      string                      `json:"WorkspaceID,omitempty"`
	ID               string                      `json:"ID,omitempty"`
	Name             *string                     `json:"-"`
	Description      *string                     `json:"-"`
	Transport        *string                     `json:"-"`
	Endpoint         *string                     `json:"-"`
	Headers          map[string]string           `json:"-"`
	Env              map[string]string           `json:"-"`
	Command          *string                     `json:"-"`
	Args             []string                    `json:"-"`
	AuthType         *string                     `json:"-"`
	CredentialConfig *V1MCPCredentialInputParams `json:"-"`
	ToolAllowlist    []string                    `json:"-"`
	ToolDenylist     []string                    `json:"-"`
	ToolPrefix       *string                     `json:"-"`
	Timeout          *int64                      `json:"-"`
	Status           *string                     `json:"-"`
	Source           *string                     `json:"-"`
}

type V1MCPDeleteParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	ID          string `json:"ID,omitempty"`
}

type V1MCPTestConnectionParams struct {
	WorkspaceID      string                      `json:"WorkspaceID,omitempty"`
	Transport        string                      `json:"Transport,omitempty"`
	Endpoint         string                      `json:"URL,omitempty"`
	Headers          map[string]string           `json:"Headers,omitempty"`
	Env              map[string]string           `json:"Env,omitempty"`
	Command          string                      `json:"Command,omitempty"`
	Args             []string                    `json:"Args,omitempty"`
	AuthType         string                      `json:"AuthType,omitempty"`
	CredentialConfig *V1MCPCredentialInputParams `json:"CredentialConfig,omitempty"`
	Timeout          int64                       `json:"Timeout,omitempty"`
}

type V1MCPTestConnectionResult struct {
	Success   bool        `json:"Success"`
	Error     string      `json:"Error,omitempty"`
	ToolCount int32       `json:"ToolCount"`
	Tools     []V1MCPTool `json:"Tools,omitempty"`
}

type V1MCPTool struct {
	Name        string `json:"Name"`
	Description string `json:"Description,omitempty"`
}

// V1MCPCredentialInputParams 对应服务端 CredentialConfig 入参，用于在创建 / 更新 / 测试连接 MCP
// 时同时声明所需凭证；服务端会据此创建或绑定 credential provider。
type V1MCPCredentialInputParams struct {
	Name         string                          `json:"Name,omitempty"`
	Description  string                          `json:"Description,omitempty"`
	Source       string                          `json:"Source,omitempty"`
	ProviderType string                          `json:"ProviderType,omitempty"`
	Config       json.RawMessage                 `json:"Config,omitempty"`
	Secrets      []V1CredentialSecretInputParams `json:"Secrets,omitempty"`
}

// V1CredentialSecretInputParams 对应服务端 CredentialSecretInput；密钥实际值字段名为
// SecretValue（与服务端 IDL 对齐）。
type V1CredentialSecretInputParams struct {
	SecretID    string `json:"SecretID,omitempty"`
	KeyName     string `json:"KeyName,omitempty"`
	Description string `json:"Description,omitempty"`
	SecretType  string `json:"SecretType,omitempty"`
	SecretValue string `json:"SecretValue,omitempty"`
}
