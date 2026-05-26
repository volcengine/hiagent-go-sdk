package v1

import (
	"context"
	"errors"
	"fmt"

	"github.com/volcengine/hiagent-go-sdk/hibot/internal/request"
	"github.com/volcengine/hiagent-go-sdk/hibot/internal/version"
)

type AgentsService struct{ client *Client }

func (s *AgentsService) New(ctx context.Context, params V1AgentNewParams) (*V1Agent, error) {
	envID := params.EnvID
	if envID == "" {
		env, err := s.client.Environments.Default(ctx, V1EnvironmentListParams{
			WorkspaceID: params.WorkspaceID,
		})
		if err != nil {
			return nil, fmt.Errorf("select environment: %w", err)
		}
		envID = env.ID
	}
	body := map[string]any{
		"WorkspaceID": params.WorkspaceID,
		"Name":        params.Name,
		"ModelID":     params.Model.ID,
		"EnvID":       envID,
	}
	if params.System != nil {
		body["SystemPrompt"] = *params.System
	}
	if resources := buildResourceInput(params.Resources); resources != nil {
		body["Resources"] = resources
	}
	if skills, mcps := buildToolBindings(params.Tools); len(skills) > 0 || len(mcps) > 0 {
		if len(skills) > 0 {
			body["Skills"] = skills
		}
		if len(mcps) > 0 {
			body["MCPs"] = mcps
		}
	}
	var result V1Agent
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "CreateAgent",
		Body:    body,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: create agent response missing ID")
	}
	result.Name = params.Name
	result.ModelID = params.Model.ID
	return &result, nil
}

func (s *AgentsService) List(ctx context.Context, params V1AgentListParams) ([]V1Agent, error) {
	var result struct {
		Items []V1Agent `json:"Items"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "ListAgents",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (s *AgentsService) Get(ctx context.Context, params V1AgentGetParams) (*V1Agent, error) {
	if params.AgentID == "" {
		return nil, errors.New("hibot: agent id is required")
	}
	var result V1Agent
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "GetAgent",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: get agent response missing ID")
	}
	return &result, nil
}

func (s *AgentsService) BatchGet(ctx context.Context, params V1AgentBatchGetParams) ([]V1Agent, error) {
	if len(params.AgentIDs) == 0 {
		return nil, errors.New("hibot: AgentIDs is required")
	}
	var result struct {
		Items []V1Agent `json:"Items"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "BatchGetAgents",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (s *AgentsService) Update(ctx context.Context, params V1AgentUpdateParams) error {
	if params.AgentID == "" {
		return errors.New("hibot: agent id is required")
	}
	body := map[string]any{
		"WorkspaceID": params.WorkspaceID,
		"AgentID":     params.AgentID,
	}
	if params.Description != nil {
		body["Description"] = *params.Description
	}
	if params.ModelID != nil {
		body["ModelID"] = *params.ModelID
	}
	if params.EnvID != nil {
		body["EnvID"] = *params.EnvID
	}
	if params.System != nil {
		body["SystemPrompt"] = *params.System
	}
	if params.Skills != nil {
		skills := make([]map[string]any, 0, len(params.Skills))
		for _, t := range params.Skills {
			if t.SkillVersionID == "" {
				continue
			}
			skills = append(skills, map[string]any{"ID": t.SkillVersionID})
		}
		body["Skills"] = skills
	}
	if params.MCPs != nil {
		mcps := make([]map[string]any, 0, len(params.MCPs))
		for _, t := range params.MCPs {
			if t.ID == "" {
				continue
			}
			mcps = append(mcps, map[string]any{"ID": t.ID, "Enabled": true})
		}
		body["MCPs"] = mcps
	}
	if params.ResetResources || len(params.Resources) > 0 {
		if resources := buildResourceInput(params.Resources); resources != nil {
			body["Resources"] = resources
		} else {
			body["Resources"] = map[string]any{}
		}
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "UpdateAgent",
		Body:    body,
	}, nil)
}

func (s *AgentsService) Delete(ctx context.Context, params V1AgentDeleteParams) error {
	if params.AgentID == "" {
		return errors.New("hibot: agent id is required")
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "DeleteAgent",
		Body:    params,
	}, nil)
}

func buildResourceInput(resources []V1ManagedAgentResourceRefParams) map[string]any {
	if len(resources) == 0 {
		return nil
	}
	resourceIDs := make([]string, 0, len(resources))
	directoryIDs := make([]string, 0, len(resources))
	for _, r := range resources {
		switch {
		case r.DirectoryID != "":
			directoryIDs = append(directoryIDs, r.DirectoryID)
		case r.ID != "":
			resourceIDs = append(resourceIDs, r.ID)
		}
	}
	out := map[string]any{}
	if len(resourceIDs) > 0 {
		out["ResourceIDs"] = resourceIDs
	}
	if len(directoryIDs) > 0 {
		out["DirectoryIDs"] = directoryIDs
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func buildToolBindings(tools []V1AgentNewParamsToolUnion) ([]map[string]any, []map[string]any) {
	skills := make([]map[string]any, 0)
	mcps := make([]map[string]any, 0)
	for _, t := range tools {
		if t.OfSkill != nil && t.OfSkill.SkillVersionID != "" {
			skills = append(skills, map[string]any{"ID": t.OfSkill.SkillVersionID})
		}
		if t.OfMCP != nil && t.OfMCP.ID != "" {
			mcps = append(mcps, map[string]any{"ID": t.OfMCP.ID, "Enabled": true})
		}
	}
	return skills, mcps
}
