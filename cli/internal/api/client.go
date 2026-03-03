package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/errfmt"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	client  *resty.Client
	baseURL string
}

type Emblem struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	AuthorID          string   `json:"author_id,omitempty"`
	AuthorName        string   `json:"author_name,omitempty"`
	Category          string   `json:"category,omitempty"`
	Tags              []string `json:"tags,omitempty"`
	License           string   `json:"license"`
	RepositoryURL     string   `json:"repository_url,omitempty"`
	HomepageURL       string   `json:"homepage_url,omitempty"`
	LatestVersion     string   `json:"latest_version"`
	Downloads         int      `json:"downloads_count"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
	SecurityAdvisory  string   `json:"security_advisory,omitempty"`
	SecuritySeverity  string   `json:"security_severity,omitempty"`
}

type Key struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type KeyCreateRequest struct {
	Name        string `json:"name"`
	ExpiresDays *int   `json:"expires_days,omitempty"`
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

func NewClientWithBaseURL(baseURL string) *Client {
	client := resty.New()
	client.SetTimeout(30 * time.Second)

	return &Client{
		client:  client,
		baseURL: baseURL,
	}
}

func (c *Client) SetToken(token string) {
	c.client.SetAuthToken(token)
}

func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
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
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return nil, errfmt.ConnectionError(c.baseURL, err)
		}
		if strings.Contains(err.Error(), "timeout") {
			return nil, errfmt.NewDetailedError(err).
				WithReason("Request timed out").
				WithSuggestion("Try again or check your network connection")
		}
		return nil, errfmt.NetworkError(err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		return nil, errfmt.APIError(resp.StatusCode(), errResp.Error)
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
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return nil, errfmt.ConnectionError(c.baseURL, err)
		}
		return nil, errfmt.NetworkError(err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		return nil, errfmt.APIError(resp.StatusCode(), errResp.Error)
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
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return nil, errfmt.ConnectionError(c.baseURL, err)
		}
		return nil, errfmt.NetworkError(err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		if resp.StatusCode() == 404 {
			return nil, errfmt.EmblemNotFoundError(name)
		}
		return nil, errfmt.APIError(resp.StatusCode(), errResp.Error)
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
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return nil, errfmt.ConnectionError(c.baseURL, err)
		}
		return nil, errfmt.NetworkError(err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		if resp.StatusCode() == 404 {
			return nil, errfmt.NewDetailedError(fmt.Errorf("version '%s' not found for emblem '%s'", version, name)).
				WithSuggestion(fmt.Sprintf("Try: ely pull %s (to get latest version)", name))
		}
		return nil, errfmt.APIError(resp.StatusCode(), errResp.Error)
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
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return nil, errfmt.ConnectionError(c.baseURL, err)
		}
		return nil, errfmt.NetworkError(err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		return nil, errfmt.APIError(resp.StatusCode(), errResp.Error)
	}

	var emblem Emblem
	if err := json.Unmarshal(resp.Body(), &emblem); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &emblem, nil
}

func (c *Client) ListKeys() ([]Key, error) {
	resp, err := c.client.R().Get(c.baseURL + "/api/keys")
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return nil, errfmt.ConnectionError(c.baseURL, err)
		}
		return nil, errfmt.NetworkError(err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		if resp.StatusCode() == 401 {
			return nil, errfmt.AuthRequiredError("API_KEY")
		}
		return nil, errfmt.APIError(resp.StatusCode(), errResp.Error)
	}

	var keys []Key
	if err := json.Unmarshal(resp.Body(), &keys); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return keys, nil
}

func (c *Client) CreateKey(name string, expiresAt *time.Time) (*Key, error) {
	body := map[string]interface{}{
		"name": name,
	}

	if expiresAt != nil {
		expiresDays := int(time.Until(*expiresAt).Hours() / 24)
		if expiresDays > 0 {
			body["expires_days"] = expiresDays
		}
	}

	resp, err := c.client.R().
		SetBody(body).
		Post(c.baseURL + "/api/keys")
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return nil, errfmt.ConnectionError(c.baseURL, err)
		}
		return nil, errfmt.NetworkError(err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		if resp.StatusCode() == 401 {
			return nil, errfmt.AuthRequiredError("API_KEY")
		}
		return nil, errfmt.APIError(resp.StatusCode(), errResp.Error)
	}

	var key Key
	if err := json.Unmarshal(resp.Body(), &key); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &key, nil
}

func (c *Client) GetKey(id string) (*Key, error) {
	resp, err := c.client.R().Get(c.baseURL + "/api/keys/" + id)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return nil, errfmt.ConnectionError(c.baseURL, err)
		}
		return nil, errfmt.NetworkError(err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		if resp.StatusCode() == 401 {
			return nil, errfmt.AuthRequiredError("API_KEY")
		}
		return nil, errfmt.APIError(resp.StatusCode(), errResp.Error)
	}

	var key Key
	if err := json.Unmarshal(resp.Body(), &key); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &key, nil
}

func (c *Client) DeleteKey(id string) error {
	resp, err := c.client.R().Delete(c.baseURL + "/api/keys/" + id)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return errfmt.ConnectionError(c.baseURL, err)
		}
		return errfmt.NetworkError(err)
	}

	if resp.IsError() {
		var errResp ErrorResponse
		json.Unmarshal(resp.Body(), &errResp)
		if resp.StatusCode() == 401 {
			return errfmt.AuthRequiredError("API_KEY")
		}
		return errfmt.APIError(resp.StatusCode(), errResp.Error)
	}

	return nil
}
