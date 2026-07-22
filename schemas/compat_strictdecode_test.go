package schemas

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// strictDecodeFixtureDir holds the ADR-0017 §9 strict-decode compat fixtures
// (P3-E2-S01): every safety-bearing authored resource (Config,
// RulesetBinding, MergePolicy — the P3-E1 schema set) must reject unknown
// fields, duplicate collection IDs, and unknown enum values at decode time,
// and must still accept the corresponding clean fixture (REQ-P3-E2-S01-04).
const strictDecodeFixtureDir = "testdata/compat/strict-decode"

// strictDecodeCase names one (resource type, fixture file) pair under
// strictDecodeFixtureDir for a given resource's compiled schema.
type strictDecodeCase struct {
	schema  *jsonschema.Schema
	fixture string
}

func readFixture(t *testing.T, relPath string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(strictDecodeFixtureDir, relPath)) //nolint:gosec // relPath is a hardcoded test-fixture literal, never external input.
	if err != nil {
		t.Fatalf("read fixture %s: %v", relPath, err)
	}
	return string(raw)
}

// TestStrictDecode proves REQ-P3-E2-S01-01..04: unknown field, duplicate ID,
// and unknown enum fixtures are rejected for every safety-bearing resource
// type, and the corresponding valid fixture is not — so a validator that
// simply rejects everything cannot pass this suite either.
func TestStrictDecode(t *testing.T) {
	t.Run("unknown_field", func(t *testing.T) {
		// REQ-P3-E2-S01-01: a field absent from the schema is rejected with
		// a positioned error naming it — every fixture plants "bogusField".
		for name, tc := range map[string]strictDecodeCase{
			"Config":         {schema: ConfigSchema, fixture: "config/unknown-field.json"},
			"RulesetBinding": {schema: RulesetBindingSchema, fixture: "ruleset-binding/unknown-field.json"},
			"MergePolicy":    {schema: MergePolicySchema, fixture: "merge-policy/unknown-field.json"},
		} {
			t.Run(name, func(t *testing.T) {
				doc := readFixture(t, tc.fixture)
				err := validateJSON(tc.schema, doc)
				if err == nil {
					t.Fatalf("expected %s fixture with an unknown field to be rejected", name)
				}
				if !strings.Contains(err.Error(), "bogusField") {
					t.Fatalf("expected a positioned error naming %q, got: %v", "bogusField", err)
				}
			})
		}
	})

	t.Run("duplicate_id", func(t *testing.T) {
		// REQ-P3-E2-S01-02: two named-collection elements sharing one ID are
		// rejected even when they differ only in an unrelated field, so a
		// naive last-write-wins merge could not hide the duplicate.
		for name, tc := range map[string]strictDecodeCase{
			"Config":         {schema: ConfigSchema, fixture: "config/duplicate-id.json"},
			"RulesetBinding": {schema: RulesetBindingSchema, fixture: "ruleset-binding/duplicate-id.json"},
			"MergePolicy":    {schema: MergePolicySchema, fixture: "merge-policy/duplicate-id.json"},
		} {
			t.Run(name, func(t *testing.T) {
				doc := readFixture(t, tc.fixture)
				if err := validateJSON(tc.schema, doc); err == nil {
					t.Fatalf("expected %s fixture with a duplicate collection ID to be rejected", name)
				}
			})
		}
	})

	t.Run("unknown_enum", func(t *testing.T) {
		// REQ-P3-E2-S01-03: an enum value outside the frozen set is
		// rejected rather than coerced. RulesetBinding has no multi-value
		// enum field today, so its fixture exercises the "kind" discriminator
		// (a single-value enum expressed as "const") as the equivalent proof
		// point for this resource type.
		for name, tc := range map[string]strictDecodeCase{
			"Config":         {schema: ConfigSchema, fixture: "config/unknown-enum.json"},
			"RulesetBinding": {schema: RulesetBindingSchema, fixture: "ruleset-binding/unknown-enum.json"},
			"MergePolicy":    {schema: MergePolicySchema, fixture: "merge-policy/unknown-enum.json"},
		} {
			t.Run(name, func(t *testing.T) {
				doc := readFixture(t, tc.fixture)
				if err := validateJSON(tc.schema, doc); err == nil {
					t.Fatalf("expected %s fixture with an unknown enum value to be rejected", name)
				}
			})
		}
	})

	t.Run("valid", func(t *testing.T) {
		// REQ-P3-E2-S01-04: the same three rules must not reject a clean
		// fixture — asserted both directions so no rule is "satisfied" by a
		// validator that rejects everything.
		for name, tc := range map[string]strictDecodeCase{
			"Config":         {schema: ConfigSchema, fixture: "config/valid.json"},
			"RulesetBinding": {schema: RulesetBindingSchema, fixture: "ruleset-binding/valid.json"},
			"MergePolicy":    {schema: MergePolicySchema, fixture: "merge-policy/valid.json"},
		} {
			t.Run(name, func(t *testing.T) {
				doc := readFixture(t, tc.fixture)
				if err := validateJSON(tc.schema, doc); err != nil {
					t.Fatalf("expected %s valid fixture to decode cleanly, got: %v", name, err)
				}
			})
		}
	})
}
