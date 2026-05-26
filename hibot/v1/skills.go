package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/volcengine/hiagent-go-sdk/hibot/internal/request"
	"github.com/volcengine/hiagent-go-sdk/hibot/internal/version"
)

type SkillsService struct{ client *Client }

func (s *SkillsService) New(ctx context.Context, params V1SkillNewParams) (*V1SkillVersion, error) {
	if params.Source == "" {
		params.Source = "manual"
	}
	var result struct {
		ID string `json:"ID"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "CreateSkill",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: create skill response missing ID")
	}
	return &V1SkillVersion{
		ID:      result.ID,
		SkillID: params.SkillID,
		Name:    params.Name,
		Version: params.Version,
	}, nil
}

func (s *SkillsService) List(ctx context.Context, params V1SkillListParams) ([]V1Skill, error) {
	var result struct {
		Items []V1Skill `json:"Items"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "ListSkills",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (s *SkillsService) Get(ctx context.Context, params V1SkillGetParams) (*V1Skill, error) {
	if params.ID == "" && params.SkillID == "" {
		return nil, errors.New("hibot: skill id or skill_id is required")
	}
	var result V1Skill
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "GetSkill",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: get skill response missing ID")
	}
	return &result, nil
}

func (s *SkillsService) Update(ctx context.Context, params V1SkillUpdateParams) error {
	if params.ID == "" && params.SkillID == "" {
		return errors.New("hibot: skill id or skill_id is required")
	}
	body := map[string]any{
		"WorkspaceID": params.WorkspaceID,
	}
	if params.ID != "" {
		body["ID"] = params.ID
	}
	if params.SkillID != "" {
		body["SkillID"] = params.SkillID
	}
	if params.Version != "" {
		body["Version"] = params.Version
	}
	if params.Description != nil {
		body["Description"] = *params.Description
	}
	if params.Source != nil {
		body["Source"] = *params.Source
	}
	if params.ArtifactID != nil {
		body["ArtifactID"] = *params.ArtifactID
	}
	if params.Enabled != nil {
		body["Enabled"] = *params.Enabled
	}
	if params.NewVersion != nil {
		body["NewVersion"] = *params.NewVersion
	}
	if params.SlugID != nil {
		body["SlugID"] = *params.SlugID
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "UpdateSkill",
		Body:    body,
	}, nil)
}

func (s *SkillsService) Delete(ctx context.Context, params V1SkillDeleteParams) error {
	if params.ID == "" && params.SkillID == "" {
		return errors.New("hibot: skill id or skill_id is required")
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "DeleteSkill",
		Body:    params,
	}, nil)
}

func (s *SkillsService) ListVersions(ctx context.Context, params V1SkillVersionListParams) ([]V1SkillVersion, error) {
	if params.SkillID == "" {
		return nil, errors.New("hibot: skill_id is required")
	}
	var result struct {
		Items []V1SkillVersion `json:"Items"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "ListSkillVersions",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (s *SkillsService) ResolveVersion(ctx context.Context, params V1SkillResolveVersionParams) (*V1SkillVersion, error) {
	if params.ID != "" {
		return &V1SkillVersion{ID: params.ID, Name: params.Name, Constraint: params.Constraint}, nil
	}
	skillID, err := s.resolveSkillID(ctx, params)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"WorkspaceID": params.WorkspaceID,
		"SkillID":     skillID,
	}
	var raw json.RawMessage
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "ListSkillVersions",
		Body:    body,
	}, &raw); err != nil {
		return nil, err
	}
	version := firstSkillVersion(raw)
	if version.ID == "" {
		return nil, fmt.Errorf("hibot: no skill version matched name=%q constraint=%q", params.Name, params.Constraint)
	}
	version.Name = params.Name
	version.Constraint = params.Constraint
	return &version, nil
}

func (s *SkillsService) resolveSkillID(ctx context.Context, params V1SkillResolveVersionParams) (string, error) {
	body := map[string]any{
		"WorkspaceID": params.WorkspaceID,
		"Name":        params.Name,
	}
	var result struct {
		Items []struct {
			SkillID string `json:"SkillID"`
			Name    string `json:"Name"`
		} `json:"Items"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "ListSkills",
		Body:    body,
	}, &result); err != nil {
		return "", err
	}
	for _, item := range result.Items {
		if item.Name == params.Name && item.SkillID != "" {
			return item.SkillID, nil
		}
	}
	if len(result.Items) > 0 && result.Items[0].SkillID != "" {
		return result.Items[0].SkillID, nil
	}
	return "", fmt.Errorf("hibot: skill %q not found", params.Name)
}

func firstSkillVersion(raw json.RawMessage) V1SkillVersion {
	var list struct {
		Items []V1SkillVersion `json:"Items"`
	}
	if err := json.Unmarshal(raw, &list); err == nil && len(list.Items) > 0 {
		return list.Items[0]
	}
	var single V1SkillVersion
	if err := json.Unmarshal(raw, &single); err == nil {
		return single
	}
	return V1SkillVersion{}
}
