package schemas

import "testing"

// REQ-P3-E1-S01-01: config.schema.json requires apiVersion/kind consts,
// non-empty environments[]/classes[] each with a match object, and an
// optional providers map (type + failure: enum [closed, open], ADR-0004).
func TestConfigSchema(t *testing.T) {
	t.Run("ADR-0010 config.yaml shape is valid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "Config",
			"environments": [
				{"name": "prod", "match": {"paths": ["topics/prod/**", "envs/prod/**"]}},
				{"name": "dev", "match": {"paths": ["**"]}}
			],
			"classes": [
				{"name": "kafka-topic", "match": {"paths": ["topics/**/*.yaml"]}},
				{"name": "infra-vars", "match": {"paths": ["**/*.tfvars"]}}
			],
			"providers": {
				"author": {"type": "builtin/gitlab-groups"},
				"quota": {"type": "http", "url": "https://quota.example.com/api/v1/lookup", "failure": "closed"}
			}
		}`
		if err := validateJSON(ConfigSchema, doc); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
	})

	t.Run("minimal config with no providers is valid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "Config",
			"environments": [{"name": "dev", "match": {"paths": ["**"]}}],
			"classes": [{"name": "kafka-topic", "match": {"paths": ["topics/**"]}}]
		}`
		if err := validateJSON(ConfigSchema, doc); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
	})

	t.Run("provider failure open is valid (opt-in, ADR-0004)", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "Config",
			"environments": [{"name": "dev", "match": {"paths": ["**"]}}],
			"classes": [{"name": "kafka-topic", "match": {"paths": ["topics/**"]}}],
			"providers": {"quota": {"type": "http", "failure": "open"}}
		}`
		if err := validateJSON(ConfigSchema, doc); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
	})

	t.Run("wrong apiVersion is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1beta1",
			"kind": "Config",
			"environments": [{"name": "dev", "match": {"paths": ["**"]}}],
			"classes": [{"name": "kafka-topic", "match": {"paths": ["topics/**"]}}]
		}`
		if err := validateJSON(ConfigSchema, doc); err == nil {
			t.Fatal("expected invalid apiVersion to fail validation")
		}
	})

	t.Run("empty environments is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "Config",
			"environments": [],
			"classes": [{"name": "kafka-topic", "match": {"paths": ["topics/**"]}}]
		}`
		if err := validateJSON(ConfigSchema, doc); err == nil {
			t.Fatal("expected empty environments[] to fail validation")
		}
	})

	t.Run("empty classes is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "Config",
			"environments": [{"name": "dev", "match": {"paths": ["**"]}}],
			"classes": []
		}`
		if err := validateJSON(ConfigSchema, doc); err == nil {
			t.Fatal("expected empty classes[] to fail validation")
		}
	})

	t.Run("adversarial: classes entry missing match is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "Config",
			"environments": [{"name": "dev", "match": {"paths": ["**"]}}],
			"classes": [{"name": "kafka-topic"}]
		}`
		if err := validateJSON(ConfigSchema, doc); err == nil {
			t.Fatal("expected classes[] entry without match to fail validation")
		}
	})

	t.Run("adversarial: providers entry with unknown key is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "Config",
			"environments": [{"name": "dev", "match": {"paths": ["**"]}}],
			"classes": [{"name": "kafka-topic", "match": {"paths": ["topics/**"]}}],
			"providers": {"quota": {"type": "http", "bogus": "field"}}
		}`
		if err := validateJSON(ConfigSchema, doc); err == nil {
			t.Fatal("expected providers entry with an unknown key to fail validation")
		}
	})

	t.Run("adversarial: providers entry with invalid failure enum is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "Config",
			"environments": [{"name": "dev", "match": {"paths": ["**"]}}],
			"classes": [{"name": "kafka-topic", "match": {"paths": ["topics/**"]}}],
			"providers": {"quota": {"type": "http", "failure": "sometimes"}}
		}`
		if err := validateJSON(ConfigSchema, doc); err == nil {
			t.Fatal("expected providers entry with an invalid failure enum to fail validation")
		}
	})

	t.Run("adversarial: unknown top-level field is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "Config",
			"environments": [{"name": "dev", "match": {"paths": ["**"]}}],
			"classes": [{"name": "kafka-topic", "match": {"paths": ["topics/**"]}}],
			"bogus": true
		}`
		if err := validateJSON(ConfigSchema, doc); err == nil {
			t.Fatal("expected unknown top-level field to fail validation")
		}
	})
}
