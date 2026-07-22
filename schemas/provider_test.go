package schemas

import "testing"

// REQ-P3-E1-S03-01: request.schema.json/response.schema.json are promoted
// unchanged from hack/spikes/provider/ to schemas/provider/v1alpha1/ (byte
// content unchanged — verified separately by diffing against the pre-move
// commit); this package's own validators compile the promoted files with
// their $id left unchanged.
func TestProviderSchemasCompile(t *testing.T) {
	if ProviderRequestSchema == nil {
		t.Fatal("ProviderRequestSchema did not compile")
	}
	if ProviderResponseSchema == nil {
		t.Fatal("ProviderResponseSchema did not compile")
	}
}

// REQ-P3-E1-S03-02: the promoted response.schema.json still enforces (via
// if/then) that state: resolved requires value + expiresAt, and any other
// state forbids value — distinct machine states, never a silently absent
// key (ADR-0017 §6). Per the frozen Spike C shape, the schema itself does
// NOT structurally require a "reason" string on non-resolved states (that
// is host/lint-layer convention, not a JSON-Schema-enforced rule) — flagged
// to the operator INBOX as a promotion-time observation, not changed here,
// since REQ-P3-E1-S03-01 requires the promoted bytes to be unchanged.
func TestProviderResponseStates(t *testing.T) {
	const fact = `{
		"name": "groups",
		"declaration": {"type": "string", "cardinality": "set", "subject": "user", "sensitive": false, "maxAge": "1h"},
		"subject": {"kind": "user", "id": "alice"},
		"observedAt": "2026-07-21T12:00:00Z"
	}`

	t.Run("resolved requires value and expiresAt", func(t *testing.T) {
		doc := `{
			"apiVersion": "provider.assent.dev/v1alpha1",
			"kind": "FactResponse",
			"queryId": "q-1",
			"facts": [` + withState(fact, `"state": "resolved", "value": ["team-a"], "expiresAt": "2026-07-21T13:00:00Z"`) + `]
		}`
		if err := validateJSON(ProviderResponseSchema, doc); err != nil {
			t.Fatalf("expected valid resolved fact, got: %v", err)
		}
	})

	t.Run("adversarial: resolved without expiresAt is invalid", func(t *testing.T) {
		doc := `{
			"apiVersion": "provider.assent.dev/v1alpha1",
			"kind": "FactResponse",
			"queryId": "q-1",
			"facts": [` + withState(fact, `"state": "resolved", "value": ["team-a"]`) + `]
		}`
		if err := validateJSON(ProviderResponseSchema, doc); err == nil {
			t.Fatal("expected resolved fact without expiresAt to fail validation")
		}
	})

	for _, state := range []string{"unavailable", "invalid", "expired"} {
		t.Run("adversarial: "+state+" with a value is invalid", func(t *testing.T) {
			doc := `{
				"apiVersion": "provider.assent.dev/v1alpha1",
				"kind": "FactResponse",
				"queryId": "q-1",
				"facts": [` + withState(fact, `"state": "`+state+`", "value": ["team-a"], "reason": "boom"`) + `]
			}`
			if err := validateJSON(ProviderResponseSchema, doc); err == nil {
				t.Fatalf("expected %s fact carrying a value to fail validation", state)
			}
		})

		t.Run(state+" without value is valid (reason optional at the schema level)", func(t *testing.T) {
			doc := `{
				"apiVersion": "provider.assent.dev/v1alpha1",
				"kind": "FactResponse",
				"queryId": "q-1",
				"facts": [` + withState(fact, `"state": "`+state+`", "reason": "boom"`) + `]
			}`
			if err := validateJSON(ProviderResponseSchema, doc); err != nil {
				t.Fatalf("expected valid %s fact, got: %v", state, err)
			}
		})
	}

	t.Run("adversarial: missing fact key never happens — facts is minItems 1", func(t *testing.T) {
		const doc = `{
			"apiVersion": "provider.assent.dev/v1alpha1",
			"kind": "FactResponse",
			"queryId": "q-1",
			"facts": []
		}`
		if err := validateJSON(ProviderResponseSchema, doc); err == nil {
			t.Fatal("expected empty facts[] to fail validation")
		}
	})
}

// withState splices extra fields (state/value/expiresAt/reason) into the
// shared fact fixture above, closing over its trailing brace.
func withState(fact, extra string) string {
	return fact[:len(fact)-1] + `, ` + extra + `}`
}

// REQ-P3-E1-S03-01: a request built from the shared FactQuery shape still
// validates unchanged against the promoted request.schema.json.
func TestProviderRequestSchema(t *testing.T) {
	const doc = `{
		"apiVersion": "provider.assent.dev/v1alpha1",
		"kind": "FactQuery",
		"queryId": "q-1",
		"asOf": "2026-07-21T12:00:00Z",
		"subject": {"kind": "user", "id": "alice"},
		"outputs": ["groups"],
		"projections": {"values": [{"pointer": "/owner", "old": "team-a", "new": "team-b"}]}
	}`
	if err := validateJSON(ProviderRequestSchema, doc); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}
