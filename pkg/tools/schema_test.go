package tools

import (
	"encoding/json"
	"testing"
)

func TestGenerateSchema(t *testing.T) {
	type SimpleArgs struct {
		Query string `json:"query" description:"The search query"`
		Limit int    `json:"limit" description:"Max results"`
	}

	fn := func(args SimpleArgs) {}

	schema, err := GenerateSchema(fn)
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	// openai.FunctionParameters is map[string]interface{}
	m := map[string]interface{}(schema)

	if m["type"] != "object" {
		t.Errorf("Expected type object, got %v", m["type"])
	}

	props, ok := m["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("properties is not a map")
	}

	if len(props) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(props))
	}

	queryProp := props["query"].(map[string]interface{})
	if queryProp["type"] != "string" {
		t.Errorf("query type mismatch")
	}
	if queryProp["description"] != "The search query" {
		t.Errorf("query description mismatch")
	}

	limitProp := props["limit"].(map[string]interface{})
	if limitProp["type"] != "integer" {
		t.Errorf("limit type mismatch")
	}

	required, ok := m["required"].([]string)
	if !ok {
		t.Fatalf("required is not a slice")
	}
	if len(required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(required))
	}
}

func TestGenerateSchema_Slice(t *testing.T) {
	type SliceArgs struct {
		Tags []string `json:"tags"`
	}
	fn := func(args SliceArgs) {}

	schema, err := GenerateSchema(fn)
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	// marshal to check json structure easiest
	b, _ := json.Marshal(schema)
	t.Logf("Schema: %s", string(b))

	m := map[string]interface{}(schema)
	props := m["properties"].(map[string]interface{})
	tags := props["tags"].(map[string]interface{})

	if tags["type"] != "array" {
		t.Errorf("Expected array type")
	}
	items := tags["items"].(map[string]interface{})
	if items["type"] != "string" {
		t.Errorf("Expected items type string")
	}
}

func TestGenerateSchema_Invalid(t *testing.T) {
	// Not a function
	_, err := GenerateSchema("not a func")
	if err == nil {
		t.Error("Expected error for non-func")
	}

	// No args
	_, err = GenerateSchema(func() {})
	if err == nil {
		t.Error("Expected error for no args")
	}

	// Non-struct arg
	_, err = GenerateSchema(func(i int) {})
	if err == nil {
		t.Error("Expected error for non-struct arg")
	}
}
