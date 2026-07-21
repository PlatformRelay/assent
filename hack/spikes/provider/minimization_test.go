package provider

import (
	"strings"
	"testing"
)

// REQ-P2-E3-S02-02: a provider declaring values.pointers [/owner] receives
// ONLY the /owner projection when the change touches /owner and /secretRef;
// requesting full old/new without the explicit trusted capability is refused
// at config load.
func TestMinimization(t *testing.T) {
	t.Run("declared projection only", func(t *testing.T) {
		cfg, err := LoadProviderConfig([]byte(`{
			"name": "toy-groups",
			"requests": {"values": {"pointers": ["/owner"]}}
		}`))
		if err != nil {
			t.Fatalf("config load: %v", err)
		}

		change := map[string]ValueChange{
			"/owner":     {Old: "team-a", New: "team-b"},
			"/secretRef": {Old: "vault/old", New: "vault/new"},
		}
		q := BuildQuery(cfg, "q-min-1", fixedAsOf, Subject{Kind: "user", ID: "alice"}, []string{"groups"}, change)

		if err := ValidateRequest(mustCanonicalJSON(t, q)); err != nil {
			t.Fatalf("built request invalid against request.schema.json: %v", err)
		}
		if len(q.Projections.Values) != 1 || q.Projections.Values[0].Pointer != "/owner" {
			t.Fatalf("projections = %+v, want exactly the declared /owner", q.Projections.Values)
		}
		raw := string(mustCanonicalJSON(t, q))
		for _, undeclared := range []string{"/secretRef", "vault/old", "vault/new"} {
			if strings.Contains(raw, undeclared) {
				t.Errorf("undeclared change content %q leaked into the request:\n%s", undeclared, raw)
			}
		}
	})

	t.Run("full content refused without trusted capability", func(t *testing.T) {
		_, err := LoadProviderConfig([]byte(`{
			"name": "greedy",
			"requests": {"fullContent": true}
		}`))
		if err == nil {
			t.Fatal("config load accepted fullContent without the trusted capability — must refuse")
		}
	})

	t.Run("full content allowed with explicit trusted capability", func(t *testing.T) {
		_, err := LoadProviderConfig([]byte(`{
			"name": "trusted",
			"requests": {"fullContent": true},
			"capabilities": ["trusted-full-content"]
		}`))
		if err != nil {
			t.Fatalf("config load refused despite explicit capability: %v", err)
		}
	})
}
