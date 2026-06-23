package fleet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"sigs.k8s.io/yaml"
)

type Validator struct {
	schema *jsonschema.Schema
}

func NewValidator(schemaPath string) (*Validator, error) {
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)

	contents, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}

	var schemaDocument any
	if err := json.Unmarshal(contents, &schemaDocument); err != nil {
		return nil, err
	}

	if err := compiler.AddResource(schemaPath, schemaDocument); err != nil {
		return nil, err
	}

	schema, err := compiler.Compile(schemaPath)
	if err != nil {
		return nil, err
	}

	return &Validator{schema: schema}, nil
}

func (v *Validator) ValidateFile(path string) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	jsonContents, err := yaml.YAMLToJSON(contents)
	if err != nil {
		return fmt.Errorf("convert YAML to JSON: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonContents))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("decode JSON document: %w", err)
	}

	if err := v.schema.Validate(value); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}
	return nil
}
