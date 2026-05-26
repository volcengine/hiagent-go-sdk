package v1

import "encoding/json"

type V1Model struct {
	ID               string            `json:"ID"`
	Name             string            `json:"Name,omitempty"`
	Type             string            `json:"Type,omitempty"`
	Provider         string            `json:"Provider,omitempty"`
	Spec             string            `json:"Spec,omitempty"`
	ModelName        string            `json:"ModelName,omitempty"`
	Description      string            `json:"Description,omitempty"`
	CreateUserName   string            `json:"CreateUserName,omitempty"`
	CreateTime       string            `json:"CreateTime,omitempty"`
	DeleteAt         string            `json:"DeleteAt,omitempty"`
	TenantId         string            `json:"TenantId,omitempty"`
	UpdateUserName   string            `json:"UpdateUserName,omitempty"`
	UpdateTime       string            `json:"UpdateTime,omitempty"`
	Status           string            `json:"Status,omitempty"`
	FeaturesConfig   json.RawMessage   `json:"FeaturesConfig,omitempty"`
	Property         json.RawMessage   `json:"Property,omitempty"`
	CredentialSchema json.RawMessage   `json:"CredentialSchema,omitempty"`
	Credential       map[string]string `json:"Credential,omitempty"`
}

type V1ModelGetParams struct {
	WorkspaceID string   `json:"WorkspaceID,omitempty"`
	ID          string   `json:"-"`
	IDs         []string `json:"IDs,omitempty"`
	// 以下字段用于按非 ID 维度查询（例如按 base 模型名 ModelName 查询自定义实例）：
	// 当 ID/IDs 都为空、且至少一个过滤项非空时，SDK 自动改走 ListModel + 客户端过滤。
	Name      string `json:"-"`
	ModelName string `json:"-"`
	Provider  string `json:"-"`
	Type      string `json:"-"`
	Spec      string `json:"-"`
}

type V1ModelListParams struct {
	WorkspaceID string       `json:"WorkspaceID,omitempty"`
	Name        string       `json:"-"`
	Page        *V1PageInput `json:"Page,omitempty"`
	SortBy      string       `json:"SortBy,omitempty"`
	SortOrder   string       `json:"SortOrder,omitempty"`
}

type V1ModelList struct {
	Items []V1Model `json:"Items"`
	Total int32     `json:"Total,omitempty"`
}

type V1ModelNewParams struct {
	WorkspaceID      string            `json:"WorkspaceID,omitempty"`
	ID               string            `json:"ID,omitempty"`
	Name             string            `json:"Name,omitempty"`
	Description      string            `json:"Description,omitempty"`
	Type             string            `json:"Type,omitempty"`
	Provider         string            `json:"Provider,omitempty"`
	Spec             string            `json:"Spec,omitempty"`
	ModelName        string            `json:"ModelName,omitempty"`
	FeaturesConfig   json.RawMessage   `json:"FeaturesConfig,omitempty"`
	Property         json.RawMessage   `json:"Property,omitempty"`
	CredentialSchema json.RawMessage   `json:"CredentialSchema,omitempty"`
	Credential       map[string]string `json:"Credential,omitempty"`
}

type V1ModelUpdateParams struct {
	WorkspaceID      string            `json:"WorkspaceID,omitempty"`
	ID               string            `json:"ID,omitempty"`
	Type             string            `json:"Type,omitempty"`
	Description      string            `json:"Description,omitempty"`
	Provider         string            `json:"Provider,omitempty"`
	Spec             string            `json:"Spec,omitempty"`
	ModelName        string            `json:"ModelName,omitempty"`
	FeaturesConfig   json.RawMessage   `json:"FeaturesConfig,omitempty"`
	Property         json.RawMessage   `json:"Property,omitempty"`
	CredentialSchema json.RawMessage   `json:"CredentialSchema,omitempty"`
	Credential       map[string]string `json:"Credential,omitempty"`
}

type V1ModelDeleteParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	ID          string `json:"ID,omitempty"`
}

type V1ModelProvider struct {
	ID               string          `json:"ID"`
	Type             string          `json:"Type,omitempty"`
	Provider         string          `json:"Provider,omitempty"`
	ModelName        string          `json:"ModelName,omitempty"`
	FeaturesConfig   json.RawMessage `json:"FeaturesConfig,omitempty"`
	Property         json.RawMessage `json:"Property,omitempty"`
	CredentialSchema json.RawMessage `json:"CredentialSchema,omitempty"`
	CreateUserName   string          `json:"CreateUserName,omitempty"`
	CreateTime       string          `json:"CreateTime,omitempty"`
	UpdateUserName   string          `json:"UpdateUserName,omitempty"`
	UpdateTime       string          `json:"UpdateTime,omitempty"`
	TenantId         string          `json:"TenantId,omitempty"`
}

type V1ModelProviderListParams struct {
	WorkspaceID string       `json:"WorkspaceID,omitempty"`
	Provider    string       `json:"-"`
	Type        string       `json:"-"`
	ModelName   string       `json:"-"`
	Features    []string     `json:"-"`
	Page        *V1PageInput `json:"Page,omitempty"`
	SortBy      string       `json:"SortBy,omitempty"`
	SortOrder   string       `json:"SortOrder,omitempty"`
}

type V1ModelProviderList struct {
	Items []V1ModelProvider `json:"Models"`
	Total int32             `json:"Total,omitempty"`
}

type V1ModelProviderGetParams struct {
	WorkspaceID string   `json:"WorkspaceID,omitempty"`
	IDs         []string `json:"IDs,omitempty"`
}

type V1ModelProviderCredentialSchemaParams struct {
	WorkspaceID string   `json:"WorkspaceID,omitempty"`
	Provider    string   `json:"Provider,omitempty"`
	Spec        string   `json:"Spec,omitempty"`
	Type        string   `json:"Type,omitempty"`
	Features    []string `json:"Features,omitempty"`
}

type V1ManagedAgentModelConfigParams struct {
	ID string `json:"ID,omitempty"`
}
