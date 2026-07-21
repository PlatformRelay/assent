package celspike_test

import (
	"math"
	"strings"
	"testing"

	celspike "github.com/PlatformRelay/assent/hack/spikes/cel"
)

func TestCoercion(t *testing.T) {
	env, err := celspike.NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	trueVal := true
	falseVal := false

	cases := []celspike.CoercionCase{
		{
			Name:     "yaml-int-vs-float-archetype-partitions",
			Source:   "yaml",
			YAML:     "old: 12\nnew: 16.0\n", // archetype base=12; float form as JSON/YAML 12.0
			OldKey:   "old",
			NewKey:   "new",
			Expr:     "new >= old",
			WantBool: &trueVal,
			Note:     "CrossTypeNumericComparisons: YAML int vs float must be true, never silent false",
		},
		{
			Name:     "yaml-int-vs-int-bounded-change",
			Source:   "yaml",
			YAML:     "old: 12\nnew: 16\n", // examples/archetypes/bounded-change
			OldKey:   "old",
			NewKey:   "new",
			Expr:     "new >= old",
			WantBool: &trueVal,
		},
		{
			Name:       "yaml-quoted-numeric-string",
			Source:     "yaml",
			YAML:       "old: 12\nnew: \"12\"\n",
			OldKey:     "old",
			NewKey:     "new",
			Expr:       "new >= old",
			WantBool:   nil,
			WantErrSub: "no such overload",
			Note:       "quoted string is not a number — typed error, not silent false",
		},
		{
			Name:     "yaml-11-octal-ish-010",
			Source:   "yaml",
			YAML:     "old: 8\nnew: 010\n", // yaml.v3 still decodes 010 as octal 8
			OldKey:   "old",
			NewKey:   "new",
			Expr:     "new == old",
			WantBool: &trueVal,
			Note:     "hazard: leading-zero YAML int becomes 8; authors must use decimal",
		},
		{
			Name:       "yaml-yes-no-remain-strings",
			Source:     "yaml",
			YAML:       "flag: yes\n",
			OldKey:     "flag",
			NewKey:     "flag",
			Expr:       "new == true",
			WantBool:   nil,
			WantErrSub: "want bool",
			Note:       "yaml.v3 keeps yes/no as strings; CEL == is silent false — adapter AsCELBool rejects",
		},
		{
			Name:     "hcl-number-vs-number",
			Source:   "hcl",
			HCL:      "partitions = 16",
			OldKey:   "",
			NewKey:   "partitions",
			Expr:     "new >= old",
			WantBool: &trueVal,
			Note:     "HCL number → int64; compared to Go old=12",
		},
		{
			Name:       "hcl-string-vs-number",
			Source:     "hcl",
			HCL:        `label = "16"`,
			OldKey:     "",
			NewKey:     "label",
			Expr:       "new >= old",
			WantBool:   nil,
			WantErrSub: "no such overload",
			Note:       "HCL string must not coerce to number",
		},
		{
			Name:       "int64-overflow-float",
			Source:     "go",
			OldKey:     "old",
			NewKey:     "new",
			Expr:       "new >= old",
			WantBool:   nil,
			WantErrSub: "int64",
			Note:       "values beyond int64 range rejected by AsCELNumber before eval",
		},
		{
			Name:     "decrease-is-false-not-error",
			Source:   "yaml",
			YAML:     "old: 16\nnew: 12\n",
			OldKey:   "old",
			NewKey:   "new",
			Expr:     "new >= old",
			WantBool: &falseVal,
			Note:     "genuine false must remain false under CrossTypeNumericComparisons",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			oldV, newV, err := loadEnds(tc)
			if err != nil {
				t.Fatalf("load: %v", err)
			}

			// Overflow case: AsCELNumber must reject before CEL sees it.
			if tc.Name == "int64-overflow-float" {
				_, err := celspike.AsCELNumber(float64(math.MaxInt64) + 2048)
				if err == nil || !strings.Contains(err.Error(), tc.WantErrSub) {
					t.Fatalf("AsCELNumber overflow: err=%v want substring %q", err, tc.WantErrSub)
				}
				return
			}

			// yes/no: prove decode is string, then adapter rejects bool use (CEL == would be silent false).
			if tc.Name == "yaml-yes-no-remain-strings" {
				m, err := celspike.DecodeYAMLMap(tc.YAML)
				if err != nil {
					t.Fatalf("yaml: %v", err)
				}
				v := m["flag"]
				if _, ok := v.(string); !ok {
					t.Fatalf("expected string for yes, got %T", v)
				}
				_, err = celspike.AsCELBool(v)
				if err == nil || !strings.Contains(err.Error(), tc.WantErrSub) {
					t.Fatalf("AsCELBool: err=%v want %q", err, tc.WantErrSub)
				}
				// Document hazard: raw CEL equality is silent false.
				ast, err := celspike.CompileExpr(env, "new == true")
				if err != nil {
					t.Fatalf("compile: %v", err)
				}
				got, _, err := celspike.EvalBool(env, ast, map[string]any{"new": v, "old": false, "facts": celspike.Facts{}})
				if err != nil {
					t.Fatalf("raw CEL should not error on string==bool, got %v", err)
				}
				if got {
					t.Fatal("raw CEL string==bool returned true; expected silent false hazard")
				}
				return
			}

			oldN, err := celspike.AsCELNumber(oldV)
			if err != nil {
				if tc.WantBool != nil {
					t.Fatalf("AsCELNumber(old): %v", err)
				}
				// string/bool path may fail at AsCELNumber or at CEL — both OK if WantBool nil
				if tc.WantErrSub != "" && strings.Contains(err.Error(), "quoted numeric") {
					return
				}
			}
			newN, err := celspike.AsCELNumber(newV)
			if err != nil {
				if tc.WantBool != nil {
					t.Fatalf("AsCELNumber(new): %v", err)
				}
				if tc.WantErrSub == "no such overload" {
					// Prefer proving CEL itself errors on string↔number when AsCELNumber is bypassed.
					newN = newV
					oldN = oldV
				} else if strings.Contains(err.Error(), tc.WantErrSub) || strings.Contains(err.Error(), "quoted") {
					return
				} else {
					t.Fatalf("AsCELNumber(new): %v", err)
				}
			}

			// For HCL number case, old is fixed archetype base.
			if tc.Source == "hcl" && tc.Name == "hcl-number-vs-number" {
				oldN = int64(12)
			}
			if tc.Source == "hcl" && tc.Name == "hcl-string-vs-number" {
				oldN = int64(12)
				newN = newV // keep string
			}
			if tc.Name == "yaml-quoted-numeric-string" {
				// Bypass AsCELNumber to show raw CEL behaviour on decoded YAML types.
				oldN, newN = oldV, newV
				if s, ok := newV.(string); !ok || s != "12" {
					t.Fatalf("expected quoted string \"12\", got %#v", newV)
				}
			}

			ast, err := celspike.CompileExpr(env, tc.Expr)
			if err != nil {
				if tc.WantBool == nil {
					if tc.WantErrSub != "" && !strings.Contains(err.Error(), tc.WantErrSub) {
						t.Fatalf("compile err=%v want substring %q", err, tc.WantErrSub)
					}
					return
				}
				t.Fatalf("compile: %v", err)
			}
			got, _, err := celspike.EvalBool(env, ast, map[string]any{
				"old": oldN,
				"new": newN,
			})
			if tc.WantBool == nil {
				if err == nil {
					t.Fatalf("want error containing %q, got ok=%v", tc.WantErrSub, got)
				}
				if tc.WantErrSub != "" && !strings.Contains(err.Error(), tc.WantErrSub) {
					t.Fatalf("err=%v want substring %q", err, tc.WantErrSub)
				}
				// Critical: must not be a silent false.
				return
			}
			if err != nil {
				t.Fatalf("eval: %v", err)
			}
			if got != *tc.WantBool {
				t.Fatalf("got %v want %v (note: %s)", got, *tc.WantBool, tc.Note)
			}
		})
	}
}

func loadEnds(tc celspike.CoercionCase) (oldV, newV any, err error) {
	switch tc.Source {
	case "yaml":
		m, err := celspike.DecodeYAMLMap(tc.YAML)
		if err != nil {
			return nil, nil, err
		}
		return m[tc.OldKey], m[tc.NewKey], nil
	case "hcl":
		m, err := celspike.DecodeHCLSample(tc.HCL)
		if err != nil {
			return nil, nil, err
		}
		return int64(12), m[tc.NewKey], nil
	case "go":
		return int64(12), float64(math.MaxInt64) + 2048, nil
	default:
		return nil, nil, nil
	}
}
