package schemas

import (
	"fmt"
	"log"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/message"
)

// uniqueKeysVocabURL identifies the vendor vocabulary that backs the
// "x-uniqueKeys" keyword. It is not a JSON-Schema-standard keyword: generic
// draft 2020-12 validators ignore it as an unknown property, so schema files
// stay portable; only this package's compiler enforces it.
const uniqueKeysVocabURL = "https://assent.dev/schemas/vocab/x-unique-keys"

// uniqueKeysExt enforces that no two items of the array it annotates share
// the same composite key, projected from the property names listed in
// "x-uniqueKeys" (ADR-0017 §9: named collections are lists with mandatory
// unique IDs; source order carries no meaning without an explicit priority).
type uniqueKeysExt struct {
	props []string
}

func (u *uniqueKeysExt) Validate(ctx *jsonschema.ValidatorContext, v any) {
	arr, ok := v.([]any)
	if !ok {
		return
	}
	seen := make(map[string]int, len(arr))
	for i, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		var key strings.Builder
		for _, p := range u.props {
			fmt.Fprintf(&key, "\x00%v", obj[p])
		}
		k := key.String()
		if j, dup := seen[k]; dup {
			ctx.AddError(&duplicateKeysError{Props: u.props, Duplicates: []int{j, i}})
			return
		}
		seen[k] = i
	}
}

type duplicateKeysError struct {
	Props      []string
	Duplicates []int
}

func (*duplicateKeysError) KeywordPath() []string { return []string{"x-uniqueKeys"} }

func (e *duplicateKeysError) LocalizedString(p *message.Printer) string {
	return p.Sprintf("items at %d and %d have the same %s", e.Duplicates[0], e.Duplicates[1], strings.Join(e.Props, "+"))
}

func uniqueKeysVocabulary() *jsonschema.Vocabulary {
	metaSchema, err := jsonschema.UnmarshalJSON(strings.NewReader(`{
		"properties": {
			"x-uniqueKeys": {
				"type": "array",
				"minItems": 1,
				"items": { "type": "string", "minLength": 1 }
			}
		}
	}`))
	if err != nil {
		log.Fatalf("x-uniqueKeys metaschema: %v", err)
	}

	c := jsonschema.NewCompiler()
	if err := c.AddResource(uniqueKeysVocabURL, metaSchema); err != nil {
		log.Fatalf("x-uniqueKeys metaschema resource: %v", err)
	}
	sch, err := c.Compile(uniqueKeysVocabURL)
	if err != nil {
		log.Fatalf("x-uniqueKeys metaschema compile: %v", err)
	}

	return &jsonschema.Vocabulary{
		URL:     uniqueKeysVocabURL,
		Schema:  sch,
		Compile: compileUniqueKeys,
	}
}

func compileUniqueKeys(_ *jsonschema.CompilerContext, obj map[string]any) (jsonschema.SchemaExt, error) {
	raw, ok := obj["x-uniqueKeys"]
	if !ok {
		return nil, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, nil
	}
	props := make([]string, 0, len(items))
	for _, item := range items {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("x-uniqueKeys: expected string items, got %T", item)
		}
		props = append(props, s)
	}
	return &uniqueKeysExt{props: props}, nil
}
