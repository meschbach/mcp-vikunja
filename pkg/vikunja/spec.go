package vikunja

import (
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v2"
)

// LoadSpec loads and converts the Swagger 2.0 spec to OpenAPI 3.0
func LoadSpec(path string) (*openapi3.T, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("spec file not found: %s", path)
	}

	// Read the YAML file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	// Parse as Swagger 2.0 (YAML format)
	var swagger openapi2.T
	if err := yaml.Unmarshal(data, &swagger); err != nil {
		return nil, fmt.Errorf("failed to parse swagger spec: %w", err)
	}

	// Convert to OpenAPI 3.0 in memory
	doc3, err := openapi2conv.ToV3(&swagger)
	if err != nil {
		return nil, fmt.Errorf("failed to convert swagger to openapi3: %w", err)
	}

	return doc3, nil
}

// GetOperation returns the OpenAPI operation definition for a given path and method
func GetOperation(spec *openapi3.T, path, method string) (*openapi3.Operation, error) {
	if spec.Paths == nil {
		return nil, fmt.Errorf("no paths defined in spec")
	}

	pathItem := spec.Paths.Map()[path]
	if pathItem == nil {
		return nil, fmt.Errorf("path not found in spec: %s", path)
	}

	var operation *openapi3.Operation
	switch method {
	case "GET":
		operation = pathItem.Get
	case "POST":
		operation = pathItem.Post
	case "PUT":
		operation = pathItem.Put
	case "DELETE":
		operation = pathItem.Delete
	case "PATCH":
		operation = pathItem.Patch
	case "HEAD":
		operation = pathItem.Head
	case "OPTIONS":
		operation = pathItem.Options
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if operation == nil {
		return nil, fmt.Errorf("method %s not found for path %s", method, path)
	}

	return operation, nil
}
