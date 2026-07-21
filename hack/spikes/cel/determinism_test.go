package celspike_test

import (
	"testing"

	celspike "github.com/PlatformRelay/assent/hack/spikes/cel"
)

func TestDeterminism(t *testing.T) {
	env, err := celspike.NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	// Archetype predicates from P1-E2 fixtures (bounded-change, ownership, allow-listed).
	preds := []struct {
		name string
		expr string
		act  map[string]any
		want bool
	}{
		{
			name: "bounded-change",
			expr: "new >= old && new <= facts.quota.max_partitions",
			act: map[string]any{
				"old": int64(12),
				"new": int64(16),
				"facts": celspike.Facts{
					Quota: celspike.QuotaFacts{MaxPartitions: 24},
				},
			},
			want: true,
		},
		{
			name: "ownership",
			expr: "entry.owner in facts.author.groups",
			act: map[string]any{
				"old":   int64(12),
				"new":   int64(12),
				"entry": map[string]any{"owner": "orders-team"},
				"facts": celspike.Facts{
					Author: celspike.AuthorFacts{
						Login:  "alice",
						Groups: []string{"orders-team"},
					},
				},
			},
			want: true,
		},
		{
			name: "allow-listed-fields",
			expr: `path in ["/partitions", "/retention_hours", "/cleanup_policy"]`,
			act: map[string]any{
				"old":   int64(12),
				"new":   int64(16),
				"path":  "/partitions",
				"facts": celspike.Facts{},
			},
			want: true,
		},
		{
			name: "bounded-change-negative",
			expr: "new >= old && new <= facts.quota.max_partitions",
			act: map[string]any{
				"old": int64(12),
				"new": int64(48),
				"facts": celspike.Facts{
					Quota: celspike.QuotaFacts{MaxPartitions: 24},
				},
			},
			want: false,
		},
	}

	var maxCost uint64
	for _, p := range preds {
		t.Run(p.name, func(t *testing.T) {
			ast, err := celspike.CompileExpr(env, p.expr)
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			a, cost1, err := celspike.EvalBool(env, ast, p.act)
			if err != nil {
				t.Fatalf("eval1: %v", err)
			}
			b, cost2, err := celspike.EvalBool(env, ast, p.act)
			if err != nil {
				t.Fatalf("eval2: %v", err)
			}
			if a != b || a != p.want {
				t.Fatalf("determinism/want: first=%v second=%v want=%v", a, b, p.want)
			}
			if cost1 != cost2 {
				t.Fatalf("cost not stable: %d vs %d", cost1, cost2)
			}
			if cost1 > celspike.CostBudget {
				t.Fatalf("cost %d exceeds budget %d", cost1, celspike.CostBudget)
			}
			if cost1 > maxCost {
				maxCost = cost1
			}
			t.Logf("cost=%d budget=%d headroom=%d", cost1, celspike.CostBudget, celspike.CostBudget-cost1)
		})
	}
	t.Logf("max archetype cost=%d budget=%d headroom=%d", maxCost, celspike.CostBudget, celspike.CostBudget-maxCost)
}
