package v1

import (
	"context"
	"errors"

	"github.com/volcengine/hiagent-go-sdk/hibot/internal/request"
	"github.com/volcengine/hiagent-go-sdk/hibot/internal/version"
)

type PromptsService struct{ client *Client }

func (s *PromptsService) New(ctx context.Context, params V1PromptNewParams) (*V1Prompt, error) {
	body := map[string]any{
		"Payload": map[string]any{
			"Name":         params.Name,
			"SystemPrompt": params.Content,
		},
	}
	if params.WorkspaceID != "" {
		body["WorkspaceID"] = params.WorkspaceID
	}
	var result struct {
		ID string `json:"ID"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "CreateAgentPromptTemplate",
		Body:    body,
	}, &result); err != nil {
		return nil, err
	}
	return &V1Prompt{ID: result.ID, Name: params.Name, Content: params.Content}, nil
}

func (s *PromptsService) List(ctx context.Context, params V1PromptListParams) ([]V1Prompt, error) {
	var result struct {
		Items []V1Prompt `json:"Items"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "ListAgentPromptTemplates",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (s *PromptsService) Update(ctx context.Context, params V1PromptUpdateParams) error {
	if params.ID == "" {
		return errors.New("hibot: prompt id is required")
	}
	payload := map[string]any{}
	if params.Name != nil {
		payload["Name"] = *params.Name
	}
	if params.Content != nil {
		payload["SystemPrompt"] = *params.Content
	}
	body := map[string]any{
		"ID":      params.ID,
		"Payload": payload,
	}
	if params.WorkspaceID != "" {
		body["WorkspaceID"] = params.WorkspaceID
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "UpdateAgentPromptTemplate",
		Body:    body,
	}, nil)
}

func (s *PromptsService) Delete(ctx context.Context, params V1PromptDeleteParams) error {
	if params.ID == "" {
		return errors.New("hibot: prompt id is required")
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "DeleteAgentPromptTemplate",
		Body:    params,
	}, nil)
}
