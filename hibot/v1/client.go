package v1

import "github.com/volcengine/hiagent-go-sdk/hibot/internal/request"

// Services contains TOP service names used by V1 resources.
type Services struct {
	Server  string
	Gateway string
	Model   string
	UP      string
}

// Client is the V1 resource client.
type Client struct {
	requester *request.Client
	services  Services

	Uploads      *UploadsService
	Environments *EnvironmentsService
	Models       *ModelsService
	Prompts      *PromptsService
	Resources    *ResourcesService
	MCPs         *MCPsService
	Skills       *SkillsService
	Agents       *AgentsService
	Sessions     *SessionsService
}

func NewClient(requester *request.Client, services Services) *Client {
	c := &Client{
		requester: requester,
		services:  services,
	}
	c.Uploads = &UploadsService{client: c}
	c.Environments = &EnvironmentsService{client: c}
	c.Models = &ModelsService{client: c}
	c.Prompts = &PromptsService{client: c}
	c.Resources = newResourcesService(c)
	c.MCPs = &MCPsService{client: c}
	c.Skills = &SkillsService{client: c}
	c.Agents = &AgentsService{client: c}
	c.Sessions = newSessionsService(c)
	return c
}
