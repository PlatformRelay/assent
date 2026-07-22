// Package schemas compiles and validates the versioned JSON Schemas that are
// assent's public contract surface (ADR-0017 §7: the serialized schemas are
// the API, not the internal Go types). This package is contract-tooling, not
// engine/decision code (D-016) — it neither trips the "no engine code before
// the Phase-3 gate" rule nor the internal/ coverage gate (D-010).
package schemas

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed policy/v1alpha1/config.schema.json
var configSchemaJSON []byte

//go:embed policy/v1alpha1/ruleset-binding.schema.json
var rulesetBindingSchemaJSON []byte

//go:embed policy/v1alpha1/merge-policy.schema.json
var mergePolicySchemaJSON []byte

var (
	// ConfigSchema validates schemas/policy/v1alpha1/config.schema.json instances.
	ConfigSchema = mustCompile("config.schema.json", configSchemaJSON)
	// RulesetBindingSchema validates schemas/policy/v1alpha1/ruleset-binding.schema.json instances.
	RulesetBindingSchema = mustCompile("ruleset-binding.schema.json", rulesetBindingSchemaJSON)
	// MergePolicySchema validates schemas/policy/v1alpha1/merge-policy.schema.json instances.
	MergePolicySchema = mustCompile("merge-policy.schema.json", mergePolicySchemaJSON)
)

// newCompiler returns a compiler with this package's vendor vocabularies
// (x-uniqueKeys) registered, so every schema in this package resolves the
// same keyword semantics.
func newCompiler() *jsonschema.Compiler {
	c := jsonschema.NewCompiler()
	c.RegisterVocabulary(uniqueKeysVocabulary())
	return c
}

func compile(name string, raw []byte) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}
	c := newCompiler()
	if err := c.AddResource(name, doc); err != nil {
		return nil, fmt.Errorf("add resource %s: %w", name, err)
	}
	return c.Compile(name)
}

func mustCompile(name string, raw []byte) *jsonschema.Schema {
	sch, err := compile(name, raw)
	if err != nil {
		panic(fmt.Sprintf("compile %s: %v", name, err))
	}
	return sch
}

// validateJSON parses raw as JSON and validates it against sch.
func validateJSON(sch *jsonschema.Schema, raw string) error {
	doc, err := jsonschema.UnmarshalJSON(strings.NewReader(raw))
	if err != nil {
		return fmt.Errorf("not JSON: %w", err)
	}
	return sch.Validate(doc)
}
