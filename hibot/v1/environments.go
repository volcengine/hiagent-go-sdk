package v1

import (
	"context"
	"errors"
	"sort"

	"github.com/volcengine/hiagent-go-sdk/hibot/internal/request"
	"github.com/volcengine/hiagent-go-sdk/hibot/internal/version"
)

type EnvironmentsService struct{ client *Client }

func (s *EnvironmentsService) New(ctx context.Context, params V1EnvironmentNewParams) (*V1Environment, error) {
	payload := map[string]any{
		"ImageType": params.ImageType,
	}
	if params.Name != "" {
		payload["Name"] = params.Name
	}
	if params.Description != "" {
		payload["Description"] = params.Description
	}
	if len(params.EnvVars) > 0 {
		payload["EnvVars"] = params.EnvVars
	}
	if params.CPULimit != "" {
		payload["CpuLimit"] = params.CPULimit
	}
	if params.MemoryLimit != "" {
		payload["MemoryLimit"] = params.MemoryLimit
	}
	if params.PVCSize != "" {
		payload["PVCSize"] = params.PVCSize
	}
	if params.DataPath != "" {
		payload["DataPath"] = params.DataPath
	}
	body := map[string]any{
		"WorkspaceID": params.WorkspaceID,
		"Payload":     payload,
	}
	var result V1Environment
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "CreateEnv",
		Body:    body,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: create environment response missing ID")
	}
	result.Name = params.Name
	result.ImageType = params.ImageType
	return &result, nil
}

func (s *EnvironmentsService) List(ctx context.Context, params V1EnvironmentListParams) ([]V1Environment, error) {
	var result struct {
		Items []V1Environment `json:"Items"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "ListEnv",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (s *EnvironmentsService) Get(ctx context.Context, params V1EnvironmentGetParams) (*V1Environment, error) {
	if params.EnvID == "" {
		return nil, errors.New("hibot: env id is required")
	}
	var result V1Environment
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "GetEnv",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: get environment response missing ID")
	}
	return &result, nil
}

func (s *EnvironmentsService) Update(ctx context.Context, params V1EnvironmentUpdateParams) error {
	if params.EnvID == "" {
		return errors.New("hibot: env id is required")
	}
	payload := map[string]any{}
	if params.Name != nil {
		payload["Name"] = *params.Name
	}
	if params.Description != nil {
		payload["Description"] = *params.Description
	}
	if params.ImageType != nil {
		payload["ImageType"] = *params.ImageType
	}
	if len(params.EnvVars) > 0 {
		payload["EnvVars"] = params.EnvVars
	}
	if params.CPULimit != nil {
		payload["CpuLimit"] = *params.CPULimit
	}
	if params.MemoryLimit != nil {
		payload["MemoryLimit"] = *params.MemoryLimit
	}
	if params.PVCSize != nil {
		payload["PVCSize"] = *params.PVCSize
	}
	if params.DataPath != nil {
		payload["DataPath"] = *params.DataPath
	}
	body := map[string]any{
		"WorkspaceID": params.WorkspaceID,
		"EnvID":       params.EnvID,
		"Payload":     payload,
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "UpdateEnv",
		Body:    body,
	}, nil)
}

func (s *EnvironmentsService) Delete(ctx context.Context, params V1EnvironmentDeleteParams) error {
	if params.EnvID == "" {
		return errors.New("hibot: env id is required")
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "DeleteEnv",
		Body:    params,
	}, nil)
}

func (s *EnvironmentsService) Default(ctx context.Context, params V1EnvironmentListParams) (*V1Environment, error) {
	items, err := s.List(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("hibot: no environment found")
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].CreatedAt == "" || items[j].CreatedAt == "" {
			return false
		}
		return items[i].CreatedAt < items[j].CreatedAt
	})
	return &items[0], nil
}
