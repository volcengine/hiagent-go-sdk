package v1

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
	WorkspaceID   string                 `json:"WorkspaceID,omitempty"`
	Name          string                 `json:"Name,omitempty"`
	Description   string                 `json:"Description,omitempty"`
	Transport     string                 `json:"Transport,omitempty"`
	Endpoint      string                 `json:"URL,omitempty"`
	Headers       map[string]string      `json:"Headers,omitempty"`
	Env           map[string]string      `json:"Env,omitempty"`
	Command       string                 `json:"Command,omitempty"`
	Args          []string               `json:"Args,omitempty"`
	AuthType      string                 `json:"AuthType,omitempty"`
	Credential    *V1CredentialRefParams `json:"Credential,omitempty"`
	ToolAllowlist []string               `json:"ToolAllowlist,omitempty"`
	ToolDenylist  []string               `json:"ToolDenylist,omitempty"`
	ToolPrefix    string                 `json:"ToolPrefix,omitempty"`
	Timeout       int64                  `json:"Timeout,omitempty"`
	Source        string                 `json:"Source,omitempty"`
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
	WorkspaceID   string            `json:"WorkspaceID,omitempty"`
	ID            string            `json:"ID,omitempty"`
	Name          *string           `json:"-"`
	Description   *string           `json:"-"`
	Transport     *string           `json:"-"`
	Endpoint      *string           `json:"-"`
	Headers       map[string]string `json:"-"`
	Env           map[string]string `json:"-"`
	Command       *string           `json:"-"`
	Args          []string          `json:"-"`
	AuthType      *string           `json:"-"`
	ToolAllowlist []string          `json:"-"`
	ToolDenylist  []string          `json:"-"`
	ToolPrefix    *string           `json:"-"`
	Timeout       *int64            `json:"-"`
	Status        *string           `json:"-"`
	Source        *string           `json:"-"`
}

type V1MCPDeleteParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	ID          string `json:"ID,omitempty"`
}

type V1MCPTestConnectionParams struct {
	WorkspaceID string                 `json:"WorkspaceID,omitempty"`
	Transport   string                 `json:"Transport,omitempty"`
	Endpoint    string                 `json:"URL,omitempty"`
	Headers     map[string]string      `json:"Headers,omitempty"`
	Env         map[string]string      `json:"Env,omitempty"`
	Command     string                 `json:"Command,omitempty"`
	Args        []string               `json:"Args,omitempty"`
	AuthType    string                 `json:"AuthType,omitempty"`
	Credential  *V1CredentialRefParams `json:"Credential,omitempty"`
	Timeout     int64                  `json:"Timeout,omitempty"`
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

type V1CredentialRefParams struct {
	Name string `json:"Name,omitempty"`
}
