package executor

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/elysium/elysium/cli/internal/emblem"
	"github.com/elysium/elysium/cli/internal/errfmt"
	"github.com/go-resty/resty/v2"
)

type Executor struct {
	definition *emblem.Definition
	client     *resty.Client
}

func New(def *emblem.Definition) *Executor {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetRedirectPolicy(resty.FlexibleRedirectPolicy(5))

	return &Executor{
		definition: def,
		client:     client,
	}
}

func (e *Executor) ListActions() []string {
	return e.definition.ListActions()
}

func (e *Executor) Execute(actionName string, params map[string]interface{}, opts FormatOptions) ([]byte, error) {
	action, err := e.definition.GetAction(actionName)
	if err != nil {
		return nil, fmt.Errorf("action not found: %w", err)
	}

	creds, err := e.definition.GetAuthCredentials()
	if err != nil {
		return nil, fmt.Errorf("authentication error: %w", err)
	}

	url := e.buildURL(action.Path, params)
	req := e.client.R()

	e.setHeaders(req, creds)
	e.setQueryParams(req, action, params)

	if action.Method == "POST" || action.Method == "PUT" || action.Method == "PATCH" {
		bodyParams := e.extractBodyParams(action, params)
		if len(bodyParams) > 0 {
			req.SetBody(bodyParams)
		}
	}

	var resp *resty.Response
	switch strings.ToUpper(action.Method) {
	case "GET":
		resp, err = req.Get(url)
	case "POST":
		resp, err = req.Post(url)
	case "PUT":
		resp, err = req.Put(url)
	case "DELETE":
		resp, err = req.Delete(url)
	case "PATCH":
		resp, err = req.Patch(url)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", action.Method)
	}

	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			return nil, errfmt.NewDetailedError(err).
				WithReason("Request timed out").
				WithContext("Timeout", "30s").
				WithSuggestion("The API is taking too long to respond. Try again or check API status.")
		}
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return nil, errfmt.ConnectionError(e.definition.BaseURL, err)
		}
		return nil, errfmt.NetworkError(err)
	}

	if resp.IsError() {
		var errResp map[string]interface{}
		json.Unmarshal(resp.Body(), &errResp)
		errMsg := ""
		if msg, ok := errResp["error"].(string); ok {
			errMsg = msg
		} else if msg, ok := errResp["message"].(string); ok {
			errMsg = msg
		} else {
			errMsg = resp.Status()
		}
		return nil, errfmt.APIError(resp.StatusCode(), errMsg)
	}

	return e.formatOutput(resp.Body(), opts)
}

func (e *Executor) buildURL(path string, params map[string]interface{}) string {
	result := path
	for key, value := range params {
		placeholder := "{" + key + "}"
		if strings.Contains(result, placeholder) {
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}
	return e.definition.BaseURL + result
}

func (e *Executor) setHeaders(req *resty.Request, creds map[string]string) {
	for key, value := range creds {
		switch key {
		case "header":
			header := value
			prefix := creds["prefix"]
			if prefix != "" {
				req.SetHeader(header, prefix+creds["value"])
			} else {
				req.SetHeader(header, creds["value"])
			}
		}
	}
}

func (e *Executor) setQueryParams(req *resty.Request, action *emblem.Action, params map[string]interface{}) {
	queryParams := make(url.Values)

	for _, param := range action.Parameters {
		if param.In == "query" {
			if val, ok := params[param.Name]; ok {
				queryParams.Add(param.Name, fmt.Sprintf("%v", val))
			} else if param.Default != nil {
				queryParams.Add(param.Name, fmt.Sprintf("%v", param.Default))
			}
		}
	}

	if len(queryParams) > 0 {
		req.SetQueryParamsFromValues(queryParams)
	}
}

func (e *Executor) extractBodyParams(action *emblem.Action, params map[string]interface{}) map[string]interface{} {
	body := make(map[string]interface{})

	for _, param := range action.Parameters {
		if param.In == "body" {
			if val, ok := params[param.Name]; ok {
				body[param.Name] = val
			} else if param.Default != nil {
				body[param.Name] = param.Default
			}
		}
	}

	return body
}

func (e *Executor) formatOutput(data []byte, opts FormatOptions) ([]byte, error) {
	return FormatOutput(data, opts)
}

func formatTable(items []interface{}) ([]byte, error) {
	if len(items) == 0 {
		return []byte("No results\n"), nil
	}

	var result strings.Builder

	switch first := items[0].(type) {
	case map[string]interface{}:
		headers := make([]string, 0, len(first))
		maxLen := make(map[string]int)

		for key := range first {
			headers = append(headers, key)
			maxLen[key] = len(key)
		}

		for _, item := range items {
			if m, ok := item.(map[string]interface{}); ok {
				for key, val := range m {
					strVal := fmt.Sprintf("%v", val)
					if len(strVal) > maxLen[key] {
						maxLen[key] = len(strVal)
					}
				}
			}
		}

		for _, h := range headers {
			result.WriteString(fmt.Sprintf("%-*s  ", maxLen[h], h))
		}
		result.WriteString("\n")

		for _, h := range headers {
			result.WriteString(strings.Repeat("-", maxLen[h]))
			result.WriteString("  ")
		}
		result.WriteString("\n")

		for _, item := range items {
			if m, ok := item.(map[string]interface{}); ok {
				for _, h := range headers {
					val := fmt.Sprintf("%v", m[h])
					result.WriteString(fmt.Sprintf("%-*s  ", maxLen[h], val))
				}
				result.WriteString("\n")
			}
		}
	}

	return []byte(result.String()), nil
}

func formatObject(obj map[string]interface{}) ([]byte, error) {
	var result strings.Builder

	result.WriteString("{\n")
	for key, val := range obj {
		result.WriteString(fmt.Sprintf("  %-20s: %v\n", key, val))
	}
	result.WriteString("}\n")

	return []byte(result.String()), nil
}

func PrintRaw(data []byte) error {
	fmt.Println(string(data))
	return nil
}

func PrintJSON(data []byte) error {
	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(data, &prettyJSON); err != nil {
		fmt.Println(string(data))
		return nil
	}

	pretty, err := json.MarshalIndent(prettyJSON, "", "  ")
	if err != nil {
		fmt.Println(string(data))
		return nil
	}

	fmt.Println(string(pretty))
	return nil
}

func ParseParams(flags map[string]string) map[string]interface{} {
	params := make(map[string]interface{})

	for key, value := range flags {
		if strings.HasPrefix(value, "[") || strings.HasPrefix(value, "{") {
			var parsed interface{}
			if err := json.Unmarshal([]byte(value), &parsed); err == nil {
				params[key] = parsed
			} else {
				params[key] = value
			}
		} else {
			params[key] = value
		}
	}

	return params
}
