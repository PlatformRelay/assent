package schemas

import "testing"

// REQ-P3-E1-S01-02: ruleset-binding.schema.json requires a named bindings[]
// collection with a mandatory unique (class, environment) pair, packs[]
// non-empty, risk.threshold a positive integer, and environment: "*" as the
// documented wildcard/default (ADR-0017 §9).
func TestRulesetBindingSchema(t *testing.T) {
	t.Run("ADR-0010 bindings.yaml shape is valid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "RulesetBinding",
			"bindings": [
				{"class": "kafka-topic", "environment": "dev", "packs": ["topics"], "risk": {"threshold": 10}},
				{"class": "kafka-topic", "environment": "prod", "packs": ["topics", "topics-strict"], "risk": {"threshold": 4}},
				{"class": "infra-vars", "environment": "*", "packs": ["tfvars"], "risk": {"threshold": 6}}
			]
		}`
		if err := validateJSON(RulesetBindingSchema, doc); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
	})

	t.Run("wrong kind is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "Bindings",
			"bindings": [{"class": "kafka-topic", "environment": "dev", "packs": ["topics"], "risk": {"threshold": 10}}]
		}`
		if err := validateJSON(RulesetBindingSchema, doc); err == nil {
			t.Fatal("expected wrong kind to fail validation")
		}
	})

	t.Run("empty bindings is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "RulesetBinding",
			"bindings": []
		}`
		if err := validateJSON(RulesetBindingSchema, doc); err == nil {
			t.Fatal("expected empty bindings[] to fail validation")
		}
	})

	t.Run("binding missing packs is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "RulesetBinding",
			"bindings": [{"class": "kafka-topic", "environment": "dev", "risk": {"threshold": 10}}]
		}`
		if err := validateJSON(RulesetBindingSchema, doc); err == nil {
			t.Fatal("expected binding without packs[] to fail validation")
		}
	})

	t.Run("binding with empty packs is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "RulesetBinding",
			"bindings": [{"class": "kafka-topic", "environment": "dev", "packs": [], "risk": {"threshold": 10}}]
		}`
		if err := validateJSON(RulesetBindingSchema, doc); err == nil {
			t.Fatal("expected binding with empty packs[] to fail validation")
		}
	})

	t.Run("binding with non-positive risk.threshold is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "RulesetBinding",
			"bindings": [{"class": "kafka-topic", "environment": "dev", "packs": ["topics"], "risk": {"threshold": 0}}]
		}`
		if err := validateJSON(RulesetBindingSchema, doc); err == nil {
			t.Fatal("expected non-positive risk.threshold to fail validation")
		}
	})

	t.Run("binding missing risk is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "RulesetBinding",
			"bindings": [{"class": "kafka-topic", "environment": "dev", "packs": ["topics"]}]
		}`
		if err := validateJSON(RulesetBindingSchema, doc); err == nil {
			t.Fatal("expected binding without risk.threshold to fail validation")
		}
	})

	t.Run("adversarial: duplicate (class, environment) pair is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "RulesetBinding",
			"bindings": [
				{"class": "kafka-topic", "environment": "dev", "packs": ["topics"], "risk": {"threshold": 10}},
				{"class": "kafka-topic", "environment": "dev", "packs": ["topics-strict"], "risk": {"threshold": 4}}
			]
		}`
		if err := validateJSON(RulesetBindingSchema, doc); err == nil {
			t.Fatal("expected duplicate (class, environment) binding pair to fail validation")
		}
	})

	t.Run("same class different environment is valid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "RulesetBinding",
			"bindings": [
				{"class": "kafka-topic", "environment": "dev", "packs": ["topics"], "risk": {"threshold": 10}},
				{"class": "kafka-topic", "environment": "prod", "packs": ["topics"], "risk": {"threshold": 4}}
			]
		}`
		if err := validateJSON(RulesetBindingSchema, doc); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
	})

	// REQ-P3-E1-S01 / ADR-0017 §2: a binding is the authored home for the named
	// obligations a (class, environment) requires — the coverage invariant that
	// merge-policy prove.obligation and evaluation-input.require both reference.
	t.Run("binding with a require obligation list is valid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "RulesetBinding",
			"bindings": [
				{"class": "kafka-topic", "environment": "prod", "packs": ["topics"], "risk": {"threshold": 4}, "require": ["ownership", "non-destructive"]}
			]
		}`
		if err := validateJSON(RulesetBindingSchema, doc); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
	})

	t.Run("binding with empty require is valid (no required obligations, vacuously covered)", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "RulesetBinding",
			"bindings": [
				{"class": "kafka-topic", "environment": "dev", "packs": ["topics"], "risk": {"threshold": 10}, "require": []}
			]
		}`
		if err := validateJSON(RulesetBindingSchema, doc); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
	})

	t.Run("adversarial: duplicate require entries are invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "RulesetBinding",
			"bindings": [
				{"class": "kafka-topic", "environment": "prod", "packs": ["topics"], "risk": {"threshold": 4}, "require": ["ownership", "ownership"]}
			]
		}`
		if err := validateJSON(RulesetBindingSchema, doc); err == nil {
			t.Fatal("expected duplicate require[] entries to fail validation")
		}
	})

	t.Run("adversarial: empty-string require entry is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "RulesetBinding",
			"bindings": [
				{"class": "kafka-topic", "environment": "prod", "packs": ["topics"], "risk": {"threshold": 4}, "require": [""]}
			]
		}`
		if err := validateJSON(RulesetBindingSchema, doc); err == nil {
			t.Fatal("expected empty-string require[] entry to fail validation")
		}
	})
}
