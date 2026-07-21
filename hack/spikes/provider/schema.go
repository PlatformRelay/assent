package provider

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed request.schema.json
var requestSchemaJSON []byte

//go:embed response.schema.json
var responseSchemaJSON []byte

var (
	requestSchema  = mustCompile("request.schema.json", requestSchemaJSON)
	responseSchema = mustCompile("response.schema.json", responseSchemaJSON)
)

func mustCompile(name string, data []byte) *jsonschema.Schema {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		panic(fmt.Sprintf("parse %s: %v", name, err))
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(name, doc); err != nil {
		panic(fmt.Sprintf("add %s: %v", name, err))
	}
	return c.MustCompile(name)
}

// ValidateRequest checks raw JSON against request.schema.json.
func ValidateRequest(raw []byte) error { return validateAgainst(requestSchema, raw) }

// ValidateResponse checks raw JSON against response.schema.json.
func ValidateResponse(raw []byte) error { return validateAgainst(responseSchema, raw) }

func validateAgainst(s *jsonschema.Schema, raw []byte) error {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("not JSON: %w", err)
	}
	return s.Validate(doc)
}
