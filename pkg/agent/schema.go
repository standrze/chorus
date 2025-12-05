package agent

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/openai/openai-go/v3"
)

// GenerateSchema inspects a function's argument (which must be a struct)
// and generates a JSON schema compatible with OpenAI's FunctionParameters.
func GenerateSchema(f interface{}) (openai.FunctionParameters, error) {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return nil, fmt.Errorf("input is not a function")
	}

	if t.NumIn() != 1 {
		return nil, fmt.Errorf("function must accept exactly one argument")
	}

	argType := t.In(0)
	if argType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("function argument must be a struct")
	}

	properties := make(map[string]interface{})
	required := []string{}

	for i := 0; i < argType.NumField(); i++ {
		field := argType.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		name := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] == "-" {
				continue
			}
			if parts[0] != "" {
				name = parts[0]
			}
		}

		desc := field.Tag.Get("description")

		propSchema, err := getTypeSchema(field.Type)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", field.Name, err)
		}

		if desc != "" {
			propSchema["description"] = desc
		}

		properties[name] = propSchema

		// Assume all fields are required unless marked optional (or pointer?)
		// For now, let's treat everything as required to be safe, or maybe check 'omitempty'
		if jsonTag != "" && strings.Contains(jsonTag, "omitempty") {
			// optional
		} else {
			required = append(required, name)
		}
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}

	return openai.FunctionParameters(schema), nil
}

func getTypeSchema(t reflect.Type) (map[string]interface{}, error) {
	switch t.Kind() {
	case reflect.String:
		return map[string]interface{}{"type": "string"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]interface{}{"type": "integer"}, nil
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{"type": "number"}, nil
	case reflect.Bool:
		return map[string]interface{}{"type": "boolean"}, nil
	case reflect.Slice, reflect.Array:
		elemSchema, err := getTypeSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"type":  "array",
			"items": elemSchema,
		}, nil
	case reflect.Struct:
		// Recursive handling for nested structs could be added here
		// For now, let's stick to primitives/slices as requested
		return nil, fmt.Errorf("nested structs are not yet supported")
	default:
		return nil, fmt.Errorf("unsupported type: %v", t.Kind())
	}
}
