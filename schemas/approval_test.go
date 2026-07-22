package schemas

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed approval/v1alpha1/approval-evidence.schema.json
var approvalEvidenceSchemaJSON []byte

// ApprovalEvidenceSchema validates schemas/approval/v1alpha1/approval-evidence.schema.json
// instances. Compiled/embedded here (not in compiler.go) per this lane's
// self-contained-ownership boundary.
var ApprovalEvidenceSchema = mustCompile("approval-evidence.schema.json", approvalEvidenceSchemaJSON)

// REQ-P3-E1-S04-01: approval-evidence.schema.json requires principal
// {id, username}, source {rule, ruleType}, eligibility
// {eligibleApproverIds: non-empty}, pins {sourceSha, targetSha,
// mergeResultDigest}, observedAt, expiresAt, and verifyingCapability — built
// from the GitLab dossier §4 evidence chain (approval_rules ->
// eligible_approvers[] -> approval_state.rules[].approved_by[]).
func TestApprovalEvidenceSchema(t *testing.T) {
	t.Run("GitLab dossier §4 evidence chain (regular rule) is valid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "ApprovalEvidence",
			"principal": {"id": "202", "username": "alice"},
			"source": {"rule": "security-review", "ruleType": "regular"},
			"eligibility": {"eligibleApproverIds": ["101", "202"]},
			"pins": {"sourceSha": "abc123", "targetSha": "def456", "mergeResultDigest": "sha256:aaaa"},
			"observedAt": "2026-07-20T10:00:00Z",
			"expiresAt": "2026-07-20T11:00:00Z",
			"verifyingCapability": "approval-rules-api"
		}`
		if err := validateJSON(ApprovalEvidenceSchema, doc); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
	})

	t.Run("CODEOWNERS-sourced rule is valid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "ApprovalEvidence",
			"principal": {"id": "303", "username": "bob", "isAuthor": false},
			"source": {"rule": "codeowners:/topics/**", "ruleType": "code_owner"},
			"eligibility": {"eligibleApproverIds": ["303", "404"]},
			"pins": {"sourceSha": "abc123", "targetSha": "def456", "mergeResultDigest": "sha256:bbbb"},
			"observedAt": "2026-07-20T10:00:00Z",
			"expiresAt": "2026-07-20T11:00:00Z",
			"verifyingCapability": "codeowners"
		}`
		if err := validateJSON(ApprovalEvidenceSchema, doc); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
	})

	t.Run("adversarial: missing pins is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "ApprovalEvidence",
			"principal": {"id": "202", "username": "alice"},
			"source": {"rule": "security-review", "ruleType": "regular"},
			"eligibility": {"eligibleApproverIds": ["202"]},
			"observedAt": "2026-07-20T10:00:00Z",
			"expiresAt": "2026-07-20T11:00:00Z",
			"verifyingCapability": "approval-rules-api"
		}`
		if err := validateJSON(ApprovalEvidenceSchema, doc); err == nil {
			t.Fatal("expected missing pins to fail validation")
		}
	})

	t.Run("adversarial: empty eligibleApproverIds is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "ApprovalEvidence",
			"principal": {"id": "202", "username": "alice"},
			"source": {"rule": "security-review", "ruleType": "regular"},
			"eligibility": {"eligibleApproverIds": []},
			"pins": {"sourceSha": "abc123", "targetSha": "def456", "mergeResultDigest": "sha256:aaaa"},
			"observedAt": "2026-07-20T10:00:00Z",
			"expiresAt": "2026-07-20T11:00:00Z",
			"verifyingCapability": "approval-rules-api"
		}`
		if err := validateJSON(ApprovalEvidenceSchema, doc); err == nil {
			t.Fatal("expected empty eligibleApproverIds (dossier §4 invalid-rule pitfall) to fail validation")
		}
	})

	t.Run("adversarial: missing principal when capability present is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "ApprovalEvidence",
			"source": {"rule": "security-review", "ruleType": "regular"},
			"eligibility": {"eligibleApproverIds": ["202"]},
			"pins": {"sourceSha": "abc123", "targetSha": "def456", "mergeResultDigest": "sha256:aaaa"},
			"observedAt": "2026-07-20T10:00:00Z",
			"expiresAt": "2026-07-20T11:00:00Z",
			"verifyingCapability": "approval-rules-api"
		}`
		if err := validateJSON(ApprovalEvidenceSchema, doc); err == nil {
			t.Fatal("expected missing principal (capability present) to fail validation")
		}
	})

	t.Run("adversarial: unknown top-level field is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "ApprovalEvidence",
			"principal": {"id": "202", "username": "alice"},
			"source": {"rule": "security-review", "ruleType": "regular"},
			"eligibility": {"eligibleApproverIds": ["202"]},
			"pins": {"sourceSha": "abc123", "targetSha": "def456", "mergeResultDigest": "sha256:aaaa"},
			"observedAt": "2026-07-20T10:00:00Z",
			"expiresAt": "2026-07-20T11:00:00Z",
			"verifyingCapability": "approval-rules-api",
			"bogus": true
		}`
		if err := validateJSON(ApprovalEvidenceSchema, doc); err == nil {
			t.Fatal("expected unknown top-level field to fail validation")
		}
	})

	t.Run("adversarial: self-approval (principal.isAuthor: true) is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "ApprovalEvidence",
			"principal": {"id": "202", "username": "alice", "isAuthor": true},
			"source": {"rule": "security-review", "ruleType": "regular"},
			"eligibility": {"eligibleApproverIds": ["202"]},
			"pins": {"sourceSha": "abc123", "targetSha": "def456", "mergeResultDigest": "sha256:aaaa"},
			"observedAt": "2026-07-20T10:00:00Z",
			"expiresAt": "2026-07-20T11:00:00Z",
			"verifyingCapability": "approval-rules-api"
		}`
		if err := validateJSON(ApprovalEvidenceSchema, doc); err == nil {
			t.Fatal("expected principal.isAuthor: true (self-approval) to fail validation")
		}
	})
}

// REQ-P3-E1-S04-02: source.ruleType is a closed enum of forge approval-rule
// kinds with no discussion/challenge-resolution value — a resolved
// discussion is acknowledgement, never authorization (ADR-0017 §3), and
// cannot be wired into require-review evidence.
func TestApprovalEvidenceExcludesDiscussion(t *testing.T) {
	t.Run("adversarial: source.ruleType discussion-resolved is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "ApprovalEvidence",
			"principal": {"id": "202", "username": "alice"},
			"source": {"rule": "resolve-all-threads", "ruleType": "discussion-resolved"},
			"eligibility": {"eligibleApproverIds": ["202"]},
			"pins": {"sourceSha": "abc123", "targetSha": "def456", "mergeResultDigest": "sha256:aaaa"},
			"observedAt": "2026-07-20T10:00:00Z",
			"expiresAt": "2026-07-20T11:00:00Z",
			"verifyingCapability": "approval-rules-api"
		}`
		if err := validateJSON(ApprovalEvidenceSchema, doc); err == nil {
			t.Fatal("expected source.ruleType: discussion-resolved to fail validation (not in the closed enum)")
		}
	})

	t.Run("no property name overlaps the publication-marker vocabulary (slot/occurrence, P3-E5)", func(t *testing.T) {
		raw := string(approvalEvidenceSchemaJSON)
		for _, banned := range []string{`"slot"`, `"occurrence"`} {
			if strings.Contains(raw, banned) {
				t.Fatalf("schema must not declare a %s field (reserved for the P3-E5 publication marker)", banned)
			}
		}
	})
}

// REQ-P3-E1-S04-03: verifyingCapability: "none" forces principal/source/
// eligibility to be structurally absent (via if/then, not merely empty) —
// D-017 B5: missing capability never auto-merges and never silently
// downgrades to a challenge/thread proxy.
func TestApprovalEvidenceCapabilityGap(t *testing.T) {
	t.Run("capability gap with no evidence fields is valid (the fail-closed shape)", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "ApprovalEvidence",
			"pins": {"sourceSha": "abc123", "targetSha": "def456", "mergeResultDigest": "sha256:aaaa"},
			"observedAt": "2026-07-20T10:00:00Z",
			"expiresAt": "2026-07-20T11:00:00Z",
			"verifyingCapability": "none"
		}`
		if err := validateJSON(ApprovalEvidenceSchema, doc); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
	})

	t.Run("adversarial: capability none with non-empty eligibility is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "ApprovalEvidence",
			"eligibility": {"eligibleApproverIds": ["202"]},
			"pins": {"sourceSha": "abc123", "targetSha": "def456", "mergeResultDigest": "sha256:aaaa"},
			"observedAt": "2026-07-20T10:00:00Z",
			"expiresAt": "2026-07-20T11:00:00Z",
			"verifyingCapability": "none"
		}`
		if err := validateJSON(ApprovalEvidenceSchema, doc); err == nil {
			t.Fatal("expected verifyingCapability: none with non-empty eligibility to fail validation")
		}
	})

	t.Run("adversarial: capability none with principal present is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "ApprovalEvidence",
			"principal": {"id": "202", "username": "alice"},
			"pins": {"sourceSha": "abc123", "targetSha": "def456", "mergeResultDigest": "sha256:aaaa"},
			"observedAt": "2026-07-20T10:00:00Z",
			"expiresAt": "2026-07-20T11:00:00Z",
			"verifyingCapability": "none"
		}`
		if err := validateJSON(ApprovalEvidenceSchema, doc); err == nil {
			t.Fatal("expected verifyingCapability: none with principal present to fail validation")
		}
	})

	t.Run("adversarial: capability none with source present is invalid", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "ApprovalEvidence",
			"source": {"rule": "security-review", "ruleType": "regular"},
			"pins": {"sourceSha": "abc123", "targetSha": "def456", "mergeResultDigest": "sha256:aaaa"},
			"observedAt": "2026-07-20T10:00:00Z",
			"expiresAt": "2026-07-20T11:00:00Z",
			"verifyingCapability": "none"
		}`
		if err := validateJSON(ApprovalEvidenceSchema, doc); err == nil {
			t.Fatal("expected verifyingCapability: none with source present to fail validation")
		}
	})

	t.Run("adversarial: empty-object principal does not satisfy the absence requirement", func(t *testing.T) {
		const doc = `{
			"apiVersion": "assent.dev/v1alpha1",
			"kind": "ApprovalEvidence",
			"principal": {},
			"pins": {"sourceSha": "abc123", "targetSha": "def456", "mergeResultDigest": "sha256:aaaa"},
			"observedAt": "2026-07-20T10:00:00Z",
			"expiresAt": "2026-07-20T11:00:00Z",
			"verifyingCapability": "none"
		}`
		if err := validateJSON(ApprovalEvidenceSchema, doc); err == nil {
			t.Fatal("expected an empty (but present) principal object under capability: none to fail validation — absence, not mere emptiness, is required")
		}
	})
}
