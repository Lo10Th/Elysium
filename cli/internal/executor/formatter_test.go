package executor

import (
	"strings"
	"testing"
)

func TestFormatJSON(t *testing.T) {
	input := []byte(`{"name":"Alice","age":30}`)

	t.Run("without pretty", func(t *testing.T) {
		out, err := formatJSON(input, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(out) != string(input) {
			t.Errorf("formatJSON(pretty=false) = %q, want %q", out, input)
		}
	})

	t.Run("with pretty", func(t *testing.T) {
		out, err := formatJSON(input, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(out), "\n") {
			t.Error("formatJSON(pretty=true) should contain newlines")
		}
		if !strings.Contains(string(out), "  ") {
			t.Error("formatJSON(pretty=true) should contain indentation")
		}
	})
}

func TestFormatJSONtoYAML(t *testing.T) {
	input := []byte(`{"name":"Alice","age":30}`)
	out, err := formatJSONtoYAML(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "name: Alice") {
		t.Errorf("YAML output should contain 'name: Alice', got: %s", out)
	}
	if !strings.Contains(string(out), "age: 30") {
		t.Errorf("YAML output should contain 'age: 30', got: %s", out)
	}
}

func TestFormatCSV(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{"id": float64(1), "name": "Alice", "price": 9.99},
		map[string]interface{}{"id": float64(2), "name": "Bob", "price": 19.99},
	}

	t.Run("all fields", func(t *testing.T) {
		out, err := formatCSV(items, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(out)
		if !strings.Contains(s, "id") {
			t.Error("CSV header should contain 'id'")
		}
		if !strings.Contains(s, "Alice") {
			t.Error("CSV data should contain 'Alice'")
		}
	})

	t.Run("selected fields", func(t *testing.T) {
		out, err := formatCSV(items, []string{"id", "name"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(out)
		if !strings.Contains(s, "id") || !strings.Contains(s, "name") {
			t.Error("CSV should contain selected fields")
		}
		if strings.Contains(s, "price") {
			t.Error("CSV should not contain 'price' when not in fields list")
		}
	})

	t.Run("empty items", func(t *testing.T) {
		out, err := formatCSV([]interface{}{}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out) != 0 {
			t.Errorf("expected empty output for empty items, got %q", out)
		}
	})
}

func TestFormatColorTable(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{"id": float64(1), "name": "T-Shirt", "stock": float64(150)},
		map[string]interface{}{"id": float64(2), "name": "Jeans", "stock": float64(30)},
		map[string]interface{}{"id": float64(3), "name": "Hat", "stock": float64(0)},
	}

	t.Run("no color", func(t *testing.T) {
		opts := FormatOptions{Format: FormatTable, NoColor: true}
		out, err := formatColorTable(items, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(out)
		if strings.Contains(s, "\033[") {
			t.Error("output should not contain ANSI codes when NoColor=true")
		}
		if !strings.Contains(s, "T-Shirt") {
			t.Error("output should contain 'T-Shirt'")
		}
		// Stock indicators should appear
		if !strings.Contains(s, "🟢") {
			t.Error("output should contain 🟢 for stock=150")
		}
		if !strings.Contains(s, "🟡") {
			t.Error("output should contain 🟡 for stock=30")
		}
		if !strings.Contains(s, "🔴") {
			t.Error("output should contain 🔴 for stock=0")
		}
	})

	t.Run("field selection", func(t *testing.T) {
		opts := FormatOptions{Format: FormatTable, NoColor: true, Fields: []string{"id", "name"}}
		out, err := formatColorTable(items, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(out)
		if strings.Contains(s, "STOCK") {
			t.Error("output should not contain 'STOCK' column when not in fields list")
		}
	})

	t.Run("empty items", func(t *testing.T) {
		opts := FormatOptions{Format: FormatTable, NoColor: true}
		out, err := formatColorTable([]interface{}{}, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(out) != "No results\n" {
			t.Errorf("expected 'No results\\n', got %q", out)
		}
	})
}

func TestStockIndicator(t *testing.T) {
	tests := []struct {
		val      interface{}
		expected string
	}{
		{float64(0), "🔴"},
		{float64(30), "🟡"},
		{float64(150), "🟢"},
		{int(0), "🔴"},
		{int(49), "🟡"},
		{int(50), "🟢"},
		{"not-a-number", ""},
	}
	for _, tt := range tests {
		got := stockIndicator(tt.val)
		if got != tt.expected {
			t.Errorf("stockIndicator(%v) = %q, want %q", tt.val, got, tt.expected)
		}
	}
}

func TestRenderTemplate(t *testing.T) {
	m := map[string]interface{}{"name": "Alice", "price": 9.99}
	result := renderTemplate("Product: {name} costs {price}", m)
	if result != "Product: Alice costs 9.99" {
		t.Errorf("renderTemplate = %q, want %q", result, "Product: Alice costs 9.99")
	}
}

func TestApplyTemplate(t *testing.T) {
	t.Run("array", func(t *testing.T) {
		data := []interface{}{
			map[string]interface{}{"name": "Alice"},
			map[string]interface{}{"name": "Bob"},
		}
		out, err := applyTemplate(data, "Name: {name}")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(out)
		if !strings.Contains(s, "Name: Alice") || !strings.Contains(s, "Name: Bob") {
			t.Errorf("applyTemplate output = %q", s)
		}
	})

	t.Run("single object", func(t *testing.T) {
		data := map[string]interface{}{"name": "Alice"}
		out, err := applyTemplate(data, "Name: {name}")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(out), "Name: Alice") {
			t.Errorf("applyTemplate output = %q", out)
		}
	})
}

func TestFormatOutput(t *testing.T) {
	jsonData := []byte(`[{"id":1,"name":"Alice","stock":100}]`)

	tests := []struct {
		name     string
		opts     FormatOptions
		contains string
	}{
		{"json raw", FormatOptions{Format: FormatJSON}, `"name":"Alice"`},
		{"json pretty", FormatOptions{Format: FormatJSON, Pretty: true}, "\"name\": \"Alice\""},
		{"yaml", FormatOptions{Format: FormatYAML}, "name: Alice"},
		{"csv", FormatOptions{Format: FormatCSV}, "Alice"},
		{"table", FormatOptions{Format: FormatTable, NoColor: true}, "Alice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := FormatOutput(jsonData, tt.opts)
			if err != nil {
				t.Fatalf("FormatOutput(%s) error: %v", tt.name, err)
			}
			if !strings.Contains(string(out), tt.contains) {
				t.Errorf("FormatOutput(%s) = %q, expected to contain %q", tt.name, out, tt.contains)
			}
		})
	}
}

func TestOrderedHeaders(t *testing.T) {
	m := map[string]interface{}{"b": 1, "a": 2, "c": 3}

	t.Run("sorted when no fields specified", func(t *testing.T) {
		headers := orderedHeaders(m, nil)
		if len(headers) != 3 || headers[0] != "a" || headers[1] != "b" || headers[2] != "c" {
			t.Errorf("orderedHeaders = %v, want [a b c]", headers)
		}
	})

	t.Run("respects specified order", func(t *testing.T) {
		headers := orderedHeaders(m, []string{"c", "a"})
		if len(headers) != 2 || headers[0] != "c" || headers[1] != "a" {
			t.Errorf("orderedHeaders = %v, want [c a]", headers)
		}
	})

	t.Run("skips non-existent fields", func(t *testing.T) {
		headers := orderedHeaders(m, []string{"a", "z"})
		if len(headers) != 1 || headers[0] != "a" {
			t.Errorf("orderedHeaders = %v, want [a]", headers)
		}
	})
}
