package v1

import (
	"context"
	"errors"
	"fmt"

	"github.com/volcengine/hiagent-go-sdk/hibot/internal/request"
	"github.com/volcengine/hiagent-go-sdk/hibot/internal/version"
)

type MCPsService struct{ client *Client }

func (s *MCPsService) New(ctx context.Context, params V1MCPNewParams) (*V1MCP, error) {
	var result struct {
		ID string `json:"ID"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "CreateMCP",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: create mcp response missing ID")
	}
	return &V1MCP{ID: result.ID, Name: params.Name, Transport: params.Transport, Endpoint: params.Endpoint}, nil
}

func (s *MCPsService) List(ctx context.Context, params V1MCPListParams) ([]V1MCP, error) {
	var result struct {
		Items []V1MCP `json:"Items"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "ListMCPs",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (s *MCPsService) Get(ctx context.Context, params V1MCPGetParams) (*V1MCP, error) {
	if params.ID == "" {
		return nil, errors.New("hibot: mcp id is required")
	}
	var result V1MCP
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "GetMCP",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: get mcp response missing ID")
	}
	return &result, nil
}

func (s *MCPsService) Update(ctx context.Context, params V1MCPUpdateParams) error {
	if params.ID == "" {
		return errors.New("hibot: mcp id is required")
	}
	body := map[string]any{
		"WorkspaceID": params.WorkspaceID,
		"ID":          params.ID,
	}
	if params.Name != nil {
		body["Name"] = *params.Name
	}
	if params.Description != nil {
		body["Description"] = *params.Description
	}
	if params.Transport != nil {
		body["Transport"] = *params.Transport
	}
	if params.Endpoint != nil {
		body["URL"] = *params.Endpoint
	}
	if params.Headers != nil {
		body["Headers"] = params.Headers
	}
	if params.Env != nil {
		body["Env"] = params.Env
	}
	if params.Command != nil {
		body["Command"] = *params.Command
	}
	if params.Args != nil {
		body["Args"] = params.Args
	}
	if params.AuthType != nil {
		body["AuthType"] = *params.AuthType
	}
	if params.ToolAllowlist != nil {
		body["ToolAllowlist"] = params.ToolAllowlist
	}
	if params.ToolDenylist != nil {
		body["ToolDenylist"] = params.ToolDenylist
	}
	if params.ToolPrefix != nil {
		body["ToolPrefix"] = *params.ToolPrefix
	}
	if params.Timeout != nil {
		body["Timeout"] = *params.Timeout
	}
	if params.Status != nil {
		body["Status"] = *params.Status
	}
	if params.Source != nil {
		body["Source"] = *params.Source
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "UpdateMCP",
		Body:    body,
	}, nil)
}

func (s *MCPsService) Delete(ctx context.Context, params V1MCPDeleteParams) error {
	if params.ID == "" {
		return errors.New("hibot: mcp id is required")
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "DeleteMCP",
		Body:    params,
	}, nil)
}

func (s *MCPsService) TestConnection(ctx context.Context, params V1MCPTestConnectionParams) (*V1MCPTestConnectionResult, error) {
	var result V1MCPTestConnectionResult
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "TestMCPConnection",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *MCPsService) Resolve(ctx context.Context, params V1MCPResolveParams) (*V1MCP, error) {
	if params.ID != "" {
		return &V1MCP{ID: params.ID, Name: params.Name}, nil
	}
	if params.Name == "" {
		return nil, errors.New("hibot: mcp id or name is required")
	}
	items, err := s.List(ctx, V1MCPListParams{
		WorkspaceID: params.WorkspaceID,
		Keyword:     params.Name,
	})
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.Name == params.Name && item.ID != "" {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("hibot: mcp %q not found", params.Name)
}
