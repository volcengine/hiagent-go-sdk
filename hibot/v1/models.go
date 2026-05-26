package v1

import (
	"context"
	"errors"

	"github.com/volcengine/hiagent-go-sdk/hibot/internal/request"
	"github.com/volcengine/hiagent-go-sdk/hibot/internal/version"
)

type ModelsService struct{ client *Client }

func (s *ModelsService) Get(ctx context.Context, params V1ModelGetParams) (*V1Model, error) {
	if params.ID != "" && len(params.IDs) == 0 {
		params.IDs = []string{params.ID}
	}
	// 当未提供 ID 但提供了其他过滤维度（Name/ModelName/Provider/Type/Spec）时，
	// 自动退化到 ListModel + 客户端过滤；这样调用方可以用 base ModelName
	// 等非 ID 字段定位工作空间内的自定义实例。
	if len(params.IDs) == 0 {
		if params.Name == "" && params.ModelName == "" && params.Provider == "" && params.Type == "" && params.Spec == "" {
			return nil, errors.New("hibot: model id is required (or provide Name/ModelName/Provider/Type/Spec)")
		}
		return s.findByFilter(ctx, params)
	}
	var result struct {
		Items []V1Model `json:"Items"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Model,
		Version: version.Model,
		Action:  "GetModel",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, errors.New("hibot: model not found")
	}
	return matchModel(result.Items, params), nil
}

// findByFilter 走 ListModel 拉取候选集再做客户端过滤；当 Name 命中时优先用网关侧 Name 过滤减少返回量。
func (s *ModelsService) findByFilter(ctx context.Context, params V1ModelGetParams) (*V1Model, error) {
	listParams := V1ModelListParams{
		WorkspaceID: params.WorkspaceID,
		Name:        params.Name,
	}
	list, err := s.List(ctx, listParams)
	if err != nil {
		return nil, err
	}
	if list == nil || len(list.Items) == 0 {
		return nil, errors.New("hibot: model not found")
	}
	if got := matchModel(list.Items, params); got != nil {
		return got, nil
	}
	return nil, errors.New("hibot: model not found matching filter")
}

// matchModel 在候选集合里挑出第一条同时满足所有非空过滤项的记录；任一字段未指定即视为通配。
func matchModel(items []V1Model, params V1ModelGetParams) *V1Model {
	for i := range items {
		m := &items[i]
		if params.Name != "" && m.Name != params.Name {
			continue
		}
		if params.ModelName != "" && m.ModelName != params.ModelName {
			continue
		}
		if params.Provider != "" && m.Provider != params.Provider {
			continue
		}
		if params.Type != "" && m.Type != params.Type {
			continue
		}
		if params.Spec != "" && m.Spec != params.Spec {
			continue
		}
		return m
	}
	if params.Name == "" && params.ModelName == "" && params.Provider == "" && params.Type == "" && params.Spec == "" {
		return &items[0]
	}
	return nil
}

func (s *ModelsService) List(ctx context.Context, params V1ModelListParams) (*V1ModelList, error) {
	body := map[string]any{}
	if params.WorkspaceID != "" {
		body["WorkspaceID"] = params.WorkspaceID
	}
	if params.Page != nil {
		body["Page"] = params.Page
	}
	if params.SortBy != "" {
		body["SortBy"] = params.SortBy
	}
	if params.SortOrder != "" {
		body["SortOrder"] = params.SortOrder
	}
	if params.Name != "" {
		body["Filter"] = map[string]any{"Name": params.Name}
	}
	var result V1ModelList
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Model,
		Version: version.Model,
		Action:  "ListModel",
		Body:    body,
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *ModelsService) New(ctx context.Context, params V1ModelNewParams) (*V1Model, error) {
	if params.Name == "" {
		return nil, errors.New("hibot: model Name is required")
	}
	if params.Type == "" {
		return nil, errors.New("hibot: model Type is required")
	}
	var result struct {
		ID string `json:"ID"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Model,
		Version: version.Model,
		Action:  "CreateModel",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: create model response missing ID")
	}
	return &V1Model{
		ID:        result.ID,
		Name:      params.Name,
		Type:      params.Type,
		Provider:  params.Provider,
		Spec:      params.Spec,
		ModelName: params.ModelName,
	}, nil
}

func (s *ModelsService) Update(ctx context.Context, params V1ModelUpdateParams) error {
	if params.ID == "" {
		return errors.New("hibot: model id is required")
	}
	if params.Type == "" {
		return errors.New("hibot: model Type is required")
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Model,
		Version: version.Model,
		Action:  "UpdateModel",
		Body:    params,
	}, nil)
}

func (s *ModelsService) Delete(ctx context.Context, params V1ModelDeleteParams) error {
	if params.ID == "" {
		return errors.New("hibot: model id is required")
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Model,
		Version: version.Model,
		Action:  "DeleteModel",
		Body:    params,
	}, nil)
}

func (s *ModelsService) ListProviders(ctx context.Context) ([]string, error) {
	var result struct {
		Providers []string `json:"Providers"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Model,
		Version: version.Model,
		Action:  "ListProvider",
		Body:    map[string]any{},
	}, &result); err != nil {
		return nil, err
	}
	return result.Providers, nil
}

func (s *ModelsService) ListModelProviders(ctx context.Context, params V1ModelProviderListParams) (*V1ModelProviderList, error) {
	body := map[string]any{}
	if params.WorkspaceID != "" {
		body["WorkspaceID"] = params.WorkspaceID
	}
	if params.Page != nil {
		body["Page"] = params.Page
	}
	if params.SortBy != "" {
		body["SortBy"] = params.SortBy
	}
	if params.SortOrder != "" {
		body["SortOrder"] = params.SortOrder
	}
	filter := map[string]any{}
	if params.Provider != "" {
		filter["Provider"] = params.Provider
	}
	if params.Type != "" {
		filter["Type"] = params.Type
	}
	if params.ModelName != "" {
		filter["ModelName"] = params.ModelName
	}
	if len(params.Features) > 0 {
		filter["Features"] = params.Features
	}
	if len(filter) > 0 {
		body["Filter"] = filter
	}
	var result V1ModelProviderList
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Model,
		Version: version.Model,
		Action:  "ListModelProvider",
		Body:    body,
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *ModelsService) GetModelProvider(ctx context.Context, params V1ModelProviderGetParams) ([]V1ModelProvider, error) {
	if len(params.IDs) == 0 {
		return nil, errors.New("hibot: provider IDs are required")
	}
	var result struct {
		Items []V1ModelProvider `json:"Items"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Model,
		Version: version.Model,
		Action:  "GetProvider",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (s *ModelsService) GetModelProviderCredentialSchema(ctx context.Context, params V1ModelProviderCredentialSchemaParams) (any, error) {
	if params.Provider == "" {
		return nil, errors.New("hibot: provider is required")
	}
	if params.Type == "" {
		return nil, errors.New("hibot: model type is required")
	}
	var result struct {
		CredentialSchema any `json:"CredentialSchema"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Model,
		Version: version.Model,
		Action:  "GetModelProviderCredentialSchema",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return result.CredentialSchema, nil
}
