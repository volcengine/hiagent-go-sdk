package v1

import "encoding/json"

type V1Environment struct {
	ID          string          `json:"ID"`
	Name        string          `json:"Name,omitempty"`
	Description string          `json:"Description,omitempty"`
	ImageType   string          `json:"ImageType,omitempty"`
	EnvVars     json.RawMessage `json:"EnvVars,omitempty"`
	CPULimit    string          `json:"CpuLimit,omitempty"`
	MemoryLimit string          `json:"MemoryLimit,omitempty"`
	PVCSize     string          `json:"PVCSize,omitempty"`
	DataPath    string          `json:"DataPath,omitempty"`
	CreatedAt   string          `json:"CreatedAt,omitempty"`
	UpdatedAt   string          `json:"UpdatedAt,omitempty"`
	CreatedBy   string          `json:"CreatedBy,omitempty"`
	UpdatedBy   string          `json:"UpdatedBy,omitempty"`
}

type V1EnvironmentNewParams struct {
	WorkspaceID string          `json:"WorkspaceID,omitempty"`
	Name        string          `json:"Name,omitempty"`
	Description string          `json:"Description,omitempty"`
	ImageType   string          `json:"ImageType,omitempty"`
	EnvVars     json.RawMessage `json:"EnvVars,omitempty"`
	CPULimit    string          `json:"CpuLimit,omitempty"`
	MemoryLimit string          `json:"MemoryLimit,omitempty"`
	PVCSize     string          `json:"PVCSize,omitempty"`
	DataPath    string          `json:"DataPath,omitempty"`
}

type V1EnvironmentListParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
}

type V1EnvironmentGetParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	EnvID       string `json:"EnvID,omitempty"`
}

type V1EnvironmentUpdateParams struct {
	WorkspaceID string          `json:"WorkspaceID,omitempty"`
	EnvID       string          `json:"EnvID,omitempty"`
	Name        *string         `json:"-"`
	Description *string         `json:"-"`
	ImageType   *string         `json:"-"`
	EnvVars     json.RawMessage `json:"-"`
	CPULimit    *string         `json:"-"`
	MemoryLimit *string         `json:"-"`
	PVCSize     *string         `json:"-"`
	DataPath    *string         `json:"-"`
}

type V1EnvironmentDeleteParams struct {
	WorkspaceID string `json:"WorkspaceID,omitempty"`
	EnvID       string `json:"EnvID,omitempty"`
}
