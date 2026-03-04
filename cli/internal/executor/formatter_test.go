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

// ─── isColorEnabled ───────────────────────────────────────────────────────────

func TestIsColorEnabled(t *testing.T) {
	t.Run("disabled when noColor option is true", func(t *testing.T) {
		// NO_COLOR env irrelevant here; the noColor flag takes priority.
		t.Setenv("NO_COLOR", "")
		if isColorEnabled(true) {
			t.Error("isColorEnabled(noColor=true) should return false")
		}
	})

	t.Run("disabled when NO_COLOR env var is set", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		if isColorEnabled(false) {
			t.Error("isColorEnabled(false) should return false when NO_COLOR is set")
		}
	})

	t.Run("enabled when no override is active", func(t *testing.T) {
		// Use t.Setenv to clear NO_COLOR for this sub-test and restore it afterwards.
		t.Setenv("NO_COLOR", "")
		if !isColorEnabled(false) {
			t.Error("isColorEnabled(false) should return true when NO_COLOR is empty")
		}
	})
}

// ─── applyColor ───────────────────────────────────────────────────────────────

func TestApplyColor(t *testing.T) {
	t.Run("color disabled returns plain text", func(t *testing.T) {
		result := applyColor("hello", ansiCyan, false)
		if result != "hello" {
			t.Errorf("applyColor(disabled) = %q, want %q", result, "hello")
		}
	})

	t.Run("color enabled wraps text with ANSI codes", func(t *testing.T) {
		result := applyColor("hello", ansiCyan, true)
		if !strings.Contains(result, "hello") {
			t.Errorf("applyColor(enabled) = %q, should contain 'hello'", result)
		}
		if !strings.Contains(result, ansiReset) {
			t.Errorf("applyColor(enabled) = %q, should contain ANSI reset sequence", result)
		}
	})
}

// ─── formatColorObject ────────────────────────────────────────────────────────

func TestFormatColorObject(t *testing.T) {
	t.Run("renders all fields without ANSI when NoColor", func(t *testing.T) {
		obj := map[string]interface{}{"name": "Widget", "price": 9.99}
		opts := FormatOptions{NoColor: true}

		out, err := formatColorObject(obj, opts)
		if err != nil {
			t.Fatalf("formatColorObject() error = %v", err)
		}
		s := string(out)
		if !strings.Contains(s, "Widget") {
			t.Errorf("formatColorObject() output = %q, want to contain 'Widget'", s)
		}
		if !strings.Contains(s, "{") || !strings.Contains(s, "}") {
			t.Errorf("formatColorObject() output should contain braces: %q", s)
		}
		if strings.Contains(s, "\033[") {
			t.Error("formatColorObject() with NoColor=true should not contain ANSI codes")
		}
	})

	t.Run("respects Fields filter", func(t *testing.T) {
		obj := map[string]interface{}{"name": "Widget", "price": 9.99, "stock": 100}
		opts := FormatOptions{NoColor: true, Fields: []string{"name", "price"}}

		out, err := formatColorObject(obj, opts)
		if err != nil {
			t.Fatalf("formatColorObject() error = %v", err)
		}
		if strings.Contains(string(out), "stock") {
			t.Error("formatColorObject() output should not contain 'stock' when not in Fields")
		}
		if !strings.Contains(string(out), "Widget") {
			t.Error("formatColorObject() output should contain 'Widget'")
		}
	})

	t.Run("emits ANSI codes when color enabled", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		obj := map[string]interface{}{"key": "value"}
		opts := FormatOptions{NoColor: false}

		out, err := formatColorObject(obj, opts)
		if err != nil {
			t.Fatalf("formatColorObject() error = %v", err)
		}
		if !strings.Contains(string(out), "\033[") {
			t.Error("formatColorObject() with color enabled should contain ANSI codes")
		}
	})
}

// ─── FormatOutput additional paths ────────────────────────────────────────────

func TestFormatOutput_Template(t *testing.T) {
	data := []byte(`[{"name":"Alice","age":30},{"name":"Bob","age":25}]`)
	opts := FormatOptions{Format: FormatTable, Template: "Name: {name}, Age: {age}"}

	out, err := FormatOutput(data, opts)
	if err != nil {
		t.Fatalf("FormatOutput(template) error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "Name: Alice") {
		t.Errorf("FormatOutput(template) = %q, want to contain 'Name: Alice'", s)
	}
	if !strings.Contains(s, "Name: Bob") {
		t.Errorf("FormatOutput(template) = %q, want to contain 'Name: Bob'", s)
	}
}

func TestFormatOutput_PlainTemplate(t *testing.T) {
	data := []byte(`{"name":"Alice"}`)
	opts := FormatOptions{Format: FormatPlain, Template: "Hello {name}"}

	out, err := FormatOutput(data, opts)
	if err != nil {
		t.Fatalf("FormatOutput(plain+template) error: %v", err)
	}
	if !strings.Contains(string(out), "Hello Alice") {
		t.Errorf("FormatOutput(plain+template) = %q, want to contain 'Hello Alice'", out)
	}
}

func TestFormatOutput_UnknownFormat(t *testing.T) {
	data := []byte(`[{"name":"Alice"}]`)
	// An unrecognised format should fall back to table rendering.
	opts := FormatOptions{Format: "unknown", NoColor: true}

	out, err := FormatOutput(data, opts)
	if err != nil {
		t.Fatalf("FormatOutput(unknown format) error: %v", err)
	}
	if !strings.Contains(string(out), "Alice") {
		t.Errorf("FormatOutput(unknown format) = %q, want to contain 'Alice'", out)
	}
}

func TestFormatOutput_CSVSingleObject(t *testing.T) {
	// A single JSON object should be treated as a one-row CSV.
	data := []byte(`{"name":"Alice","age":30}`)
	opts := FormatOptions{Format: FormatCSV}

	out, err := FormatOutput(data, opts)
	if err != nil {
		t.Fatalf("FormatOutput(CSV single object) error: %v", err)
	}
	if !strings.Contains(string(out), "Alice") {
		t.Errorf("FormatOutput(CSV single object) = %q, want to contain 'Alice'", out)
	}
}

func TestFormatOutput_CSVInvalidJSON(t *testing.T) {
	_, err := FormatOutput([]byte("not json"), FormatOptions{Format: FormatCSV})
	if err == nil {
		t.Error("FormatOutput(CSV invalid JSON) should return an error")
	}
}

func TestFormatOutput_TableInvalidJSON(t *testing.T) {
	_, err := FormatOutput([]byte("not json"), FormatOptions{Format: FormatTable})
	if err == nil {
		t.Error("FormatOutput(table invalid JSON) should return an error")
	}
}

// ─── formatJSON additional paths ──────────────────────────────────────────────

func TestFormatJSON_PrettyWithInvalidJSON(t *testing.T) {
	// Invalid input with pretty=true should return the original bytes without error.
	out, err := formatJSON([]byte("not json"), true)
	if err != nil {
		t.Fatalf("formatJSON(invalid, pretty=true) error = %v, want nil", err)
	}
	if string(out) != "not json" {
		t.Errorf("formatJSON(invalid, pretty=true) = %q, want original bytes", out)
	}
}

// ─── formatJSONtoYAML additional paths ───────────────────────────────────────

func TestFormatJSONtoYAML_InvalidJSON(t *testing.T) {
	_, err := formatJSONtoYAML([]byte("not json"))
	if err == nil {
		t.Error("formatJSONtoYAML(invalid JSON) should return an error")
	}
}

// ─── formatCSV additional paths ───────────────────────────────────────────────

func TestFormatCSV_NonMapItems(t *testing.T) {
	// Items that are not maps should produce an error.
	items := []interface{}{"string1", "string2"}
	_, err := formatCSV(items, nil)
	if err == nil {
		t.Error("formatCSV with non-map items should return an error")
	}
}

// ─── formatColorTable additional paths ───────────────────────────────────────

func TestFormatColorTable_WithWidth(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{
			"name":  "A very long product name that exceeds the column budget",
			"price": "9.99",
		},
	}
	opts := FormatOptions{Format: FormatTable, NoColor: true, Width: 30}

	out, err := formatColorTable(items, opts)
	if err != nil {
		t.Fatalf("formatColorTable(width=30) error: %v", err)
	}
	if len(out) == 0 {
		t.Error("formatColorTable(width=30) returned empty output")
	}
}

func TestFormatColorTable_NonMapFallback(t *testing.T) {
	// When items[0] is not a map, formatColorTable falls back to formatTable.
	items := []interface{}{"scalar1", "scalar2"}
	opts := FormatOptions{Format: FormatTable, NoColor: true}

	_, err := formatColorTable(items, opts)
	if err != nil {
		t.Fatalf("formatColorTable(non-map fallback) error: %v", err)
	}
}

func TestFormatColorTable_EmptyHeaders(t *testing.T) {
	// When Fields is specified but none match the object, return "No results".
	items := []interface{}{
		map[string]interface{}{"name": "Widget"},
	}
	opts := FormatOptions{Format: FormatTable, NoColor: true, Fields: []string{"nonexistent"}}

	out, err := formatColorTable(items, opts)
	if err != nil {
		t.Fatalf("formatColorTable(empty headers) error: %v", err)
	}
	if string(out) != "No results\n" {
		t.Errorf("formatColorTable(empty headers) = %q, want 'No results\\n'", out)
	}
}

// ─── stockIndicator additional types ─────────────────────────────────────────

func TestStockIndicator_Int64(t *testing.T) {
	if got := stockIndicator(int64(100)); got != "🟢" {
		t.Errorf("stockIndicator(int64(100)) = %q, want 🟢", got)
	}
}

// ─── applyTemplate scalar fallback ───────────────────────────────────────────

func TestApplyTemplate_Scalar(t *testing.T) {
	out, err := applyTemplate("a plain string", "tmpl")
	if err != nil {
		t.Fatalf("applyTemplate(scalar) error: %v", err)
	}
	if !strings.Contains(string(out), "a plain string") {
		t.Errorf("applyTemplate(scalar) = %q, want to contain the scalar value", out)
	}
}

