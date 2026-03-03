package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	client  *resty.Client
	baseURL string
}

type Emblem struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	AuthorID      string   `json:"author_id,omitempty"`
	AuthorName    string   `json:"author_name,omitempty"`
	Category      string   `json:"category,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	License       string   `json:"license"`
	RepositoryURL string   `json:"repository_url,omitempty"`
	HomepageURL   string   `json:"homepage_url,omitempty"`
	LatestVersion string   `json:"latest_version"`
	Downloads     int      `json:"downloads_count"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type EmblemVersion struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	YAMLContent string `json:"yaml_content"`
	Changelog   string `json:"changelog,omitempty"`
	PublishedAt string `json:"published_at"`
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Detail  string `json:"detail,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewClient() *Client {
	cfg := config.Get()
	client := resty.New()
	client.SetTimeout(30 * time.Second)

	return &Client{
		client:  client,
		baseURL: cfg.Registry,
	}
}

func (c *Client) SetToken(token string) {
	c.client.SetAuthToken(token)
}

func (c *Client) ListEmblems(category string, limit, offset int) ([]Emblem, error) {
	req := c.client.R().
		SetQueryParams(map[string]string{
			"limit":  fmt.Sprintf("%d", limit),
			"offset": fmt.Sprintf("%d", offset),
		})

	if category != "" {
		req.SetQueryParam("category", category)
	}

	resp, err := req.Get(c.baseURL + "/api/emblems")
	if err != nil {
		return nil, fmt.Errorf("failed to list emblems: %w", err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		return nil, fmt.Errorf("API error: %s", errResp.Error)
	}

	var emblems []Emblem
	if err := json.Unmarshal(resp.Body(), &emblems); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return emblems, nil
}

func (c *Client) SearchEmblems(query, category, sort string, limit, offset int) ([]Emblem, error) {
	req := c.client.R().
		SetQueryParams(map[string]string{
			"q":      query,
			"limit":  fmt.Sprintf("%d", limit),
			"offset": fmt.Sprintf("%d", offset),
		})

	if category != "" {
		req.SetQueryParam("category", category)
	}
	if sort != "" {
		req.SetQueryParam("sort", sort)
	}

	resp, err := req.Get(c.baseURL + "/api/emblems/search")
	if err != nil {
		return nil, fmt.Errorf("failed to search emblems: %w", err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		return nil, fmt.Errorf("API error: %s", errResp.Error)
	}

	var emblems []Emblem
	if err := json.Unmarshal(resp.Body(), &emblems); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return emblems, nil
}

func (c *Client) GetEmblem(name string) (*Emblem, error) {
	resp, err := c.client.R().Get(c.baseURL + "/api/emblems/" + name)
	if err != nil {
		return nil, fmt.Errorf("failed to get emblem: %w", err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		return nil, fmt.Errorf("API error: %s", errResp.Error)
	}

	var emblem Emblem
	if err := json.Unmarshal(resp.Body(), &emblem); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &emblem, nil
}

func (c *Client) GetEmblemVersion(name, version string) (*EmblemVersion, error) {
	resp, err := c.client.R().Get(fmt.Sprintf("%s/api/emblems/%s/%s", c.baseURL, name, version))
	if err != nil {
		return nil, fmt.Errorf("failed to get emblem version: %w", err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		return nil, fmt.Errorf("API error: %s", errResp.Error)
	}

	var ver EmblemVersion
	if err := json.Unmarshal(resp.Body(), &ver); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &ver, nil
}

func (c *Client) PublishEmblem(name, description, yamlContent, version string, category string, tags []string) (*Emblem, error) {
	body := map[string]interface{}{
		"name":         name,
		"description":  description,
		"yaml_content": yamlContent,
		"version":      version,
		"category":     category,
		"tags":         tags,
	}

	resp, err := c.client.R().
		SetBody(body).
		Post(c.baseURL + "/api/emblems")
	if err != nil {
		return nil, fmt.Errorf("failed to publish emblem: %w", err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		return nil, fmt.Errorf("API error: %s", errResp.Error)
	}

	var emblem Emblem
	if err := json.Unmarshal(resp.Body(), &emblem); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &emblem, nil
}
