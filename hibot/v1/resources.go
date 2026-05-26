package v1

import (
	"context"
	"errors"

	"github.com/volcengine/hiagent-go-sdk/hibot/internal/request"
	"github.com/volcengine/hiagent-go-sdk/hibot/internal/version"
)

type ResourcesService struct {
	client      *Client
	Directories *DirectoriesService
}

type DirectoriesService struct{ client *Client }

func newResourcesService(c *Client) *ResourcesService {
	return &ResourcesService{
		client:      c,
		Directories: &DirectoriesService{client: c},
	}
}

func (s *ResourcesService) New(ctx context.Context, params V1ResourceNewParams) (*V1Resource, error) {
	if params.Name == "" {
		return nil, errors.New("hibot: resource Name is required")
	}
	if params.BlobID == "" {
		return nil, errors.New("hibot: resource BlobID is required (call Uploads.NewBlob first)")
	}
	body := map[string]any{
		"Name":   params.Name,
		"BlobID": params.BlobID,
	}
	if params.WorkspaceID != "" {
		body["WorkspaceID"] = params.WorkspaceID
	}
	if params.DirectoryID != "" {
		body["DirectoryID"] = params.DirectoryID
	}
	var result struct {
		ID string `json:"ID"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "CreateResource",
		Body:    body,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: create resource response missing ID")
	}
	return &V1Resource{
		ID:          result.ID,
		Name:        params.Name,
		Type:        params.Type,
		DirectoryID: params.DirectoryID,
	}, nil
}

func (s *ResourcesService) List(ctx context.Context, params V1ResourceListParams) (*V1ResourceList, error) {
	var result V1ResourceList
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "ListResources",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *ResourcesService) Update(ctx context.Context, params V1ResourceUpdateParams) error {
	if params.ResourceID == "" {
		return errors.New("hibot: resource id is required")
	}
	body := map[string]any{
		"WorkspaceID": params.WorkspaceID,
		"ResourceID":  params.ResourceID,
		"Name":        params.Name,
	}
	if params.DirectoryID != nil {
		body["DirectoryID"] = *params.DirectoryID
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "UpdateResource",
		Body:    body,
	}, nil)
}

func (s *ResourcesService) Delete(ctx context.Context, params V1ResourceDeleteParams) error {
	if params.ResourceID == "" {
		return errors.New("hibot: resource id is required")
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "DeleteResource",
		Body:    params,
	}, nil)
}

func (s *ResourcesService) GetByName(ctx context.Context, params V1ResourceGetByNameParams) (*V1Resource, error) {
	if params.Name == "" {
		return nil, errors.New("hibot: resource name is required")
	}
	var result struct {
		Resource V1Resource `json:"Resource"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "GetResourceByName",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if result.Resource.ID == "" {
		return nil, errors.New("hibot: get resource by name response missing ID")
	}
	return &result.Resource, nil
}

func (s *ResourcesService) BatchGet(ctx context.Context, params V1ResourceBatchGetParams) ([]V1Resource, error) {
	if len(params.IDs) == 0 {
		return nil, errors.New("hibot: resource IDs are required")
	}
	var result struct {
		Items []V1Resource `json:"Items"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "BatchGetResources",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

// ---------- Directories ----------

func (s *DirectoriesService) New(ctx context.Context, params V1DirectoryNewParams) (*V1Directory, error) {
	if params.Name == "" {
		return nil, errors.New("hibot: directory name is required")
	}
	var result struct {
		ID string `json:"ID"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "CreateDirectory",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if result.ID == "" {
		return nil, errors.New("hibot: create directory response missing ID")
	}
	return &V1Directory{ID: result.ID, Name: params.Name, WorkspaceID: params.WorkspaceID}, nil
}

func (s *DirectoriesService) List(ctx context.Context, params V1DirectoryListParams) (*V1DirectoryList, error) {
	var result V1DirectoryList
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "ListDirectories",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *DirectoriesService) Update(ctx context.Context, params V1DirectoryUpdateParams) error {
	if params.DirectoryID == "" {
		return errors.New("hibot: directory id is required")
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "UpdateDirectory",
		Body:    params,
	}, nil)
}

func (s *DirectoriesService) Delete(ctx context.Context, params V1DirectoryDeleteParams) error {
	if params.DirectoryID == "" {
		return errors.New("hibot: directory id is required")
	}
	return s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "DeleteDirectory",
		Body:    params,
	}, nil)
}

func (s *DirectoriesService) GetByName(ctx context.Context, params V1DirectoryGetByNameParams) (*V1Directory, error) {
	if params.Name == "" {
		return nil, errors.New("hibot: directory name is required")
	}
	var result struct {
		Directory V1Directory `json:"Directory"`
	}
	if err := s.client.requester.DoAction(ctx, request.Action{
		Service: s.client.services.Server,
		Version: version.Server,
		Action:  "GetDirectoryByName",
		Body:    params,
	}, &result); err != nil {
		return nil, err
	}
	if result.Directory.ID == "" {
		return nil, errors.New("hibot: get directory by name response missing ID")
	}
	return &result.Directory, nil
}
