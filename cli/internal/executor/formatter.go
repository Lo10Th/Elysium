package executor

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// OutputFormat represents supported output formats.
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatYAML  OutputFormat = "yaml"
	FormatCSV   OutputFormat = "csv"
	FormatPlain OutputFormat = "plain"
)

// FormatOptions holds options that control output formatting.
type FormatOptions struct {
	Format   OutputFormat
	Pretty   bool
	Fields   []string
	Template string
	NoColor  bool
	Width    int
}

const (
	ansiReset  = "\033[0m"
	ansiCyan   = "\033[36m"
	ansiBold   = "\033[1m"
)

func isColorEnabled(noColor bool) bool {
	if noColor {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return true
}

func applyColor(text, color string, enabled bool) string {
	if !enabled {
		return text
	}
	return color + text + ansiReset
}

// orderedHeaders returns fields in the specified order, or all sorted keys if fields is empty.
func orderedHeaders(m map[string]interface{}, fields []string) []string {
	if len(fields) > 0 {
		result := make([]string, 0, len(fields))
		for _, f := range fields {
			if _, ok := m[f]; ok {
				result = append(result, f)
			}
		}
		return result
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// stockIndicator returns a colored circle emoji based on a numeric stock value.
func stockIndicator(val interface{}) string {
	var n float64
	switch v := val.(type) {
	case float64:
		n = v
	case int:
		n = float64(v)
	case int64:
		n = float64(v)
	default:
		return ""
	}
	if n == 0 {
		return "🔴"
	} else if n < 50 {
		return "🟡"
	}
	return "🟢"
}

// FormatOutput dispatches to the appropriate formatter based on opts.Format.
func FormatOutput(data []byte, opts FormatOptions) ([]byte, error) {
	switch opts.Format {
	case FormatJSON:
		return formatJSON(data, opts.Pretty)
	case FormatYAML:
		return formatJSONtoYAML(data)
	case FormatCSV:
		var result interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
		switch v := result.(type) {
		case []interface{}:
			return formatCSV(v, opts.Fields)
		case map[string]interface{}:
			return formatCSV([]interface{}{v}, opts.Fields)
		default:
			return data, nil
		}
	case FormatPlain, FormatTable:
		var result interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
		if opts.Template != "" {
			return applyTemplate(result, opts.Template)
		}
		switch v := result.(type) {
		case []interface{}:
			return formatColorTable(v, opts)
		case map[string]interface{}:
			return formatColorObject(v, opts)
		default:
			return data, nil
		}
	default:
		// For unknown formats fall back to table rendering.
		opts.Format = FormatTable
		return FormatOutput(data, opts)
	}
}

// formatJSON returns JSON, pretty-printed if pretty is true.
func formatJSON(data []byte, pretty bool) ([]byte, error) {
	if !pretty {
		return data, nil
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return data, nil
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return data, nil
	}
	return append(out, '\n'), nil
}

// formatJSONtoYAML converts JSON bytes to YAML output.
func formatJSONtoYAML(data []byte) ([]byte, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return yaml.Marshal(v)
}

// formatCSV converts a slice of objects to CSV bytes.
func formatCSV(items []interface{}, fields []string) ([]byte, error) {
	if len(items) == 0 {
		return []byte{}, nil
	}
	first, ok := items[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("CSV format requires an array of objects")
	}
	headers := orderedHeaders(first, fields)

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(headers); err != nil {
		return nil, err
	}
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			row := make([]string, len(headers))
			for i, h := range headers {
				row[i] = fmt.Sprintf("%v", m[h])
			}
			if err := w.Write(row); err != nil {
				return nil, err
			}
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// formatColorTable renders a slice of objects as a colored ASCII table.
func formatColorTable(items []interface{}, opts FormatOptions) ([]byte, error) {
	if len(items) == 0 {
		return []byte("No results\n"), nil
	}
	first, ok := items[0].(map[string]interface{})
	if !ok {
		return formatTable(items)
	}

	useColor := isColorEnabled(opts.NoColor)
	headers := orderedHeaders(first, opts.Fields)
	if len(headers) == 0 {
		return []byte("No results\n"), nil
	}

	// Compute max column widths.
	maxLen := make(map[string]int)
	for _, h := range headers {
		maxLen[h] = len(h)
	}
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			for _, h := range headers {
				strVal := fmt.Sprintf("%v", m[h])
				if len(strVal) > maxLen[h] {
					maxLen[h] = len(strVal)
				}
			}
		}
	}

	// Respect --width by truncating if needed.
	if opts.Width > 0 {
		totalWidth := (len(headers)-1)*2 + 2 // separators
		for _, h := range headers {
			totalWidth += maxLen[h]
		}
		if totalWidth > opts.Width {
			excess := totalWidth - opts.Width
			for i := len(headers) - 1; i >= 0 && excess > 0; i-- {
				h := headers[i]
				canTrim := maxLen[h] - len(h)
				if canTrim > 0 {
					trim := canTrim
					if trim > excess {
						trim = excess
					}
					maxLen[h] -= trim
					excess -= trim
				}
			}
		}
	}

	var result strings.Builder

	// Header row.
	for _, h := range headers {
		headerText := fmt.Sprintf("%-*s", maxLen[h], strings.ToUpper(h))
		result.WriteString(applyColor(headerText, ansiBold+ansiCyan, useColor))
		result.WriteString("  ")
	}
	result.WriteString("\n")

	// Separator row using box-drawing characters.
	for _, h := range headers {
		sep := strings.Repeat("─", maxLen[h])
		result.WriteString(applyColor(sep, ansiCyan, useColor))
		result.WriteString("  ")
	}
	result.WriteString("\n")

	// Check if any row has a "stock" field for status indicators.
	_, hasStock := first["stock"]

	// Data rows.
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			for _, h := range headers {
				val := fmt.Sprintf("%v", m[h])
				// Truncate value if wider than maxLen (rune-aware).
				runes := []rune(val)
				if len(runes) > maxLen[h] {
					val = string(runes[:maxLen[h]])
				}
				result.WriteString(fmt.Sprintf("%-*s", maxLen[h], val))
				result.WriteString("  ")
			}
			if hasStock {
				indicator := stockIndicator(m["stock"])
				if indicator != "" {
					result.WriteString(indicator)
				}
			}
			result.WriteString("\n")
		}
	}

	return []byte(result.String()), nil
}

// formatColorObject renders a single map as a key-value list with optional color.
func formatColorObject(obj map[string]interface{}, opts FormatOptions) ([]byte, error) {
	useColor := isColorEnabled(opts.NoColor)
	headers := orderedHeaders(obj, opts.Fields)

	var result strings.Builder
	result.WriteString("{\n")
	for _, key := range headers {
		keyStr := applyColor(fmt.Sprintf("  %-20s", key), ansiCyan, useColor)
		result.WriteString(fmt.Sprintf("%s: %v\n", keyStr, obj[key]))
	}
	result.WriteString("}\n")
	return []byte(result.String()), nil
}

// applyTemplate applies a simple {field} template to the given data.
// For arrays, each element is rendered on its own line.
func applyTemplate(data interface{}, tmplStr string) ([]byte, error) {
	var result strings.Builder
	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				result.WriteString(renderTemplate(tmplStr, m))
				result.WriteString("\n")
			}
		}
	case map[string]interface{}:
		result.WriteString(renderTemplate(tmplStr, v))
		result.WriteString("\n")
	default:
		result.WriteString(fmt.Sprintf("%v\n", data))
	}
	return []byte(result.String()), nil
}

// renderTemplate replaces {field} placeholders with values from m.
func renderTemplate(tmplStr string, m map[string]interface{}) string {
	result := tmplStr
	for k, v := range m {
		result = strings.ReplaceAll(result, "{"+k+"}", fmt.Sprintf("%v", v))
	}
	return result
}
