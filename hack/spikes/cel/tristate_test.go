package celspike_test

import (
	"testing"

	celspike "github.com/PlatformRelay/assent/hack/spikes/cel"
)

func TestTristate(t *testing.T) {
	env, err := celspike.NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	t.Run("vouch-pass-satisfies", func(t *testing.T) {
		// Bounded-change archetype: 12 → 16 within quota 24.
		leaf := celspike.Leaf{
			ID:  "bounded",
			CEL: "new >= old && new <= facts.quota.max_partitions",
		}
		act := map[string]any{
			"old": int64(12),
			"new": int64(16),
			"facts": celspike.Facts{
				Quota: celspike.QuotaFacts{MaxPartitions: 24},
			},
		}
		res := celspike.Walk(env, leaf, act)
		if res.State != celspike.TriPass {
			t.Fatalf("state=%s err=%v", res.State, res.Traces[0].Err)
		}
		if !celspike.VouchSatisfied(res.State) {
			t.Fatal("pass must satisfy vouch")
		}
	})

	t.Run("vouch-fail-does-not-satisfy", func(t *testing.T) {
		leaf := celspike.Leaf{ID: "bounded", CEL: "new >= old"}
		act := map[string]any{"old": int64(16), "new": int64(12), "facts": celspike.Facts{}}
		res := celspike.Walk(env, leaf, act)
		if res.State != celspike.TriFail {
			t.Fatalf("state=%s", res.State)
		}
		if celspike.VouchSatisfied(res.State) {
			t.Fatal("fail must not satisfy vouch")
		}
	})

	t.Run("adversarial-type-error-does-not-prove-obligation", func(t *testing.T) {
		// Attacker-shaped input: coerce a field to the wrong type so the
		// obligation-proving predicate errors instead of returning false.
		leaf := celspike.Leaf{
			ID: "ownership",
			// startsWith requires string; int owner → type_mismatch (in/== would silent-false).
			CEL: "entry.owner.startsWith('orders')",
		}
		act := map[string]any{
			"old":   int64(12),
			"new":   int64(12),
			"entry": map[string]any{"owner": 12345}, // type bomb: int instead of string
			"facts": celspike.Facts{
				Author: celspike.AuthorFacts{
					Login:  "alice",
					Groups: []string{"orders-team"},
				},
			},
		}
		res := celspike.Walk(env, leaf, act)
		if res.State != celspike.TriError {
			t.Fatalf("want error tri-state, got %s traces=%s", res.State, celspike.FormatTraceSummary(res.Traces))
		}
		if celspike.VouchSatisfied(res.State) {
			t.Fatal("ADR-0007: error must NOT satisfy a vouch/obligation predicate")
		}
		if res.Traces[0].ErrorClass != celspike.ClassTypeMismatch {
			t.Fatalf("want type_mismatch, got %s (%v)", res.Traces[0].ErrorClass, res.Traces[0].Err)
		}
	})
}
