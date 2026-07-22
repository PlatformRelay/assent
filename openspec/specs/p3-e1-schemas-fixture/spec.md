# P3-E1 — JSON Schemas + the strict contract fixture

**Problem**: three adversarial design reviews found that vague or missing contracts hid
safety bugs (one broad vouch covering unrelated obligations, decisions pinned to a movable
source SHA, a resolved `challenge` treated as authorization the forge never proved).
ADR-0017 §7 responds by making the **serialized JSON Schemas the public API** — Go
interfaces stay internal, there is no v1 SDK (P2-10) — so that engine code (Phase 5) is
written *against* a frozen shape, not the other way around. D-016 makes this the Phase-3
gate: **no engine code before the strict end-to-end contract fixture (§8) exists** and
validates against the same schemas as every example.
**Scope**: versioned JSON Schemas for the three authored resources (`Config`,
`RulesetBinding`, `MergePolicy` with `prove`/`onFailure`), the five runtime records
(`EvaluationInput`, `DecisionRecord`, `ReplayBundle`, `PresentationModel`,
`PublicationReceipt`), the provider request/response envelope (promoted from Spike C), the
typed `ApprovalEvidence` contract (OQ-23), the adopter test-fixture format (ADR-0014); the
frozen `assert` predicate-scope table (ADR-0013 consequences); the `assent lint` hard-error
list spec (ADR-0010 amendment) and a schema-validation CI job; the strict §8 exit-gate
fixture and one sanitized named-consumer compatibility fixture (D-017 B5).
**Non-goals**: any evaluator/engine implementation (D-016 — schemas and fixtures only, no
`internal/core` decision logic); the versioning/compatibility *rule engine* (P3-E2 turns
ADR-0017 §9 into executable compat tests — this epic only freezes the shapes those rules
will check); rollout `phase`/policy profiles/comparison records (P3-E4, ADR-0018 — OQ-21 is
explicitly out of scope here); the publication marker/reconciliation protocol (P3-E5,
ADR-0019); GitHub/Rego/CRD adapter contracts (Locked/gated — E10/E13 D-012, E14 pending
Spike D).
ADRs: 0002 §policy surface, 0003 change model, 0004 provider fail-closed, 0007 obligations
(reshaped), 0010 repo layout + lint hard-errors, 0011 ports, 0013 assert/predicate-scope,
0014 adopter test format, 0015 trust boundaries, 0016 presentation four-record split, 0017
§§1–9 (all), D-012, D-016, D-017. OQ-1 residual (apiVersion group — resolved below), OQ-8,
OQ-9, OQ-10, OQ-13, OQ-17 residual, OQ-22, OQ-23; [named-consumer-compat.md](../../../docs/planning/named-consumer-compat.md)
B5.

## Judgment calls fixed by this spec (logged to the operator INBOX as 🟡 DECIDED)

- **OQ-1 residual — apiVersion group**: `assent.dev` (already the group used in every ADR-0010/
  ADR-0016 draft sample; this spec ratifies it as frozen, not a new choice).
- **Schema file location — domain-subdirectory convention**: `schemas/<domain>/v1alpha1/
  <kind>.schema.json` at the repo root: `policy/` (Config, RulesetBinding, MergePolicy),
  `decision/` (EvaluationInput, DecisionRecord, ReplayBundle, PresentationModel,
  PublicationReceipt), `provider/` (request/response — already established by Spike C,
  unchanged), `approval/` (ApprovalEvidence), `testfixture/` (the adopter test-expectation
  format). This **adopts the convention two sibling Phase-3 lanes independently predicted**
  (P3-E4's INBOX entry: `schemas/policy/v1alpha1/…`, `schemas/decision/v1alpha1/…`,
  `schemas/comparison/v1alpha1/…`) rather than the flatter `schemas/v1alpha1/<kind>.schema.json`
  this lane drafted first — P3-E1 is the actual owner of these paths, so this spec is the
  authoritative resolution; P3-E2/P3-E4 should re-point their own predicted `Test:`/`Verify:`
  paths to match once this lane's spec lands (their INBOX entries already anticipated exactly
  this reconciliation). Root-level `schemas/` (not `docs/schemas/`) because ADR-0017 §7 makes
  these the public API, not documentation.
- **Schema self-check harness location**: `schemas/*_test.go` (package `schemas`, *not* under
  `internal/`) — a compile+validate harness is contract tooling, not engine/decision code, so
  it neither trips D-016 ("no engine code") nor the D-010 90%-coverage gate (which is scoped to
  `internal/…`), matching the precedent set by `hack/spikes/provider`. This is distinct from
  P3-E2's predicted `internal/core/schema/*_test.go` — that path is Phase-5 loader/strict-decode
  *engine* test code, sequenced after this gate closes, not this lane's schema-compiles harness.
- **Exit-gate + compatibility fixture location**: `examples/contracts/d016-strict-fixture/` and
  `examples/contracts/named-consumer-compat/`, alongside the existing `examples/archetypes/`,
  `examples/policies/`, `examples/repos/` siblings.

## P3-E1-S01 — Authored-resource schemas + the predicate-scope freeze

- **Goal**: `Config`, `RulesetBinding`, `MergePolicy` — including `prove`/`onFailure`,
  `entries:` → `EntryRef` derivation, and the four matcher domains (`files`,
  `values.pointers`, `fileEvents`, `valueChanges`, ADR-0017 §5) — exist as versioned JSON
  Schemas, and the `assert` predicate-scope table (ADR-0013 consequences) is frozen as a
  companion contract referenced by the `MergePolicy` schema description.
- **Operator input**: no (apiVersion group already ratified above; content is ADR-derived).
- **Dependencies**: none — this is the foundation slice every later story builds on.
- **Definition of done**: three schema files committed under `schemas/policy/v1alpha1/`; each
  compiles as a valid JSON Schema (draft 2020-12); the existing ADR-0010 draft samples
  validate as positive fixtures; the predicate-scope table is committed and every field name
  it lists matches a `$defs` property referenced from `MergePolicy`'s `assert` description.

Requirements:

- **REQ-P3-E1-S01-01** — Given ADR-0010's `config.yaml` shape, when
  `schemas/policy/v1alpha1/config.schema.json` is authored, then it requires
  `apiVersion: const "assent.dev/v1alpha1"`, `kind: const "Config"`, non-empty
  `environments[]` and `classes[]` (each with a `match` object), an optional `providers` map
  keyed by provider name (`type`, `failure: enum [closed, open]`, default `closed` per
  ADR-0004), and sets `additionalProperties: false` at every safety-bearing level (ADR-0017
  §9). Adversarial case: a `providers` entry with an unknown key or a `classes[]` entry
  missing `match` fails validation.
  - Test: `schemas/policy/v1alpha1/config.schema.json`
  - Verify: `go test ./schemas/... -run TestConfigSchema`
  - Level: L0
- **REQ-P3-E1-S01-02** — Given ADR-0010's `bindings.yaml` shape, when
  `schemas/policy/v1alpha1/ruleset-binding.schema.json` is authored, then `bindings[]` is a named
  collection with a **mandatory unique `class`+`environment` pair** (ADR-0017 §9: named
  collections are lists with mandatory unique IDs, source order carries no meaning without an
  explicit `priority`), each entry requires `packs[]` (non-empty) and `risk.threshold`
  (positive integer), and `environment: "*"` is accepted as the documented wildcard/default.
  Adversarial case: two bindings with the same `(class, environment)` pair fails validation.
  - Test: `schemas/policy/v1alpha1/ruleset-binding.schema.json`
  - Verify: `go test ./schemas/... -run TestRulesetBindingSchema`
  - Level: L0
- **REQ-P3-E1-S01-03** — Given ADR-0017 §§2/5, when
  `schemas/policy/v1alpha1/merge-policy.schema.json` is authored, then each `rules[]` entry has
  `match` restricted to exactly one of the four matcher-domain shapes (`files`,
  `values.pointers`, `fileEvents`, `valueChanges` — no generic `path` overload), and either
  `prove: {obligation: string, when: <assert tree>}` with `onFailure: {effect: enum [comment,
  challenge, block, require-review], code: string}` (never a bare `effect: vouch`) — `require-
  review` is satisfied only by forge-proven `ApprovalEvidence` (ADR-0017 §3, S04), never a
  bare vouch or a resolved-discussion proxy — or a non-obligation `effect`
  (`comment`/`challenge`/`block`) with no `prove`. Composition inside `when`/`assert`
  is `all`/`any`/`not` with CEL-string leaves (ADR-0013); a bare string is single-leaf
  shorthand. `entries:` (`mode: enum [document, list, map]`, `root`, `identity.pointer`) is
  present wherever a class declares governed subjects; **unkeyed `list` mode without
  `identity.pointer` fails validation** (ADR-0017 §5 — rejected at lint, not guessed).
  Adversarial case: a rule with both `prove` and effect `vouch` fails; a `list`-mode `entries`
  block without `identity.pointer` fails.
  - Test: `schemas/policy/v1alpha1/merge-policy.schema.json`
  - Verify: `go test ./schemas/... -run TestMergePolicySchema`
  - Level: L0
- **REQ-P3-E1-S01-04** — Given ADR-0013's consequences list, when the predicate-scope table is
  frozen, then `docs/adr/0013-appendix-syntax-gallery.md` or a new
  `docs/planning/predicate-scope.md` enumerates exactly `old`, `new`, `path`, `kind`, `file`,
  `entry`, `oldEntry`, `changes`, `facts`, `mr`, `env` as the closed set of top-level fields a
  CEL leaf may reference, each with its source type; the `MergePolicy` schema's `assert`/`cel`
  description references this table by name so a future `assent lint` implementation has one
  authoritative field list to check unknown-field references against (ADR-0016 §2: unknown
  fields are load-time errors, never `<no value>`).
  - Test: `docs/planning/predicate-scope.md`
  - Verify: `grep -q "oldEntry" docs/planning/predicate-scope.md && grep -q "facts" docs/planning/predicate-scope.md`
  - Level: doc

## P3-E1-S02 — Runtime record schemas + Pins (OQ-8/OQ-9)

- **Goal**: `EvaluationInput`, `DecisionRecord`, `ReplayBundle`, `PresentationModel`,
  `PublicationReceipt` (ADR-0016 §3's four-record split plus the input record) exist as
  versioned JSON Schemas, with `Pins` (tool digest + policy SHA + source/target/merge-result
  digests + fact-resolution timestamps) frozen inside `DecisionRecord`/`ReplayBundle` from day
  one (OQ-8/OQ-9).
- **Operator input**: no.
- **Dependencies**: P3-E1-S01 (obligation names, `EntryRef` shape, and matcher-domain findings
  feed `EvaluationInput`/`DecisionRecord`).
- **Definition of done**: five schema files committed under `schemas/decision/v1alpha1/`; a hand-built
  positive fixture round-trips through all five (an `EvaluationInput` instance's obligations
  and subjects reappear, correctly redacted, in the corresponding `DecisionRecord`/
  `PresentationModel`); `EvaluationInput` contains no CEL- or Rego-specific field (backend
  neutrality, keeping E11 unlocked without a contract change).
- **Note on E11 backend-neutrality**: `EvaluationInput`'s obligation, subject, and predicate
  fields are typed by *shape* (obligation name, EntryRef, per-leaf trace), never by backend
  tag — a future Rego-backed rule (E11) must be representable in the identical input record a
  CEL-backed rule uses.

Requirements:

- **REQ-P3-E1-S02-01** — Given ADR-0017's contract-ownership split, when
  `schemas/decision/v1alpha1/evaluation-input.schema.json` is authored, then it carries the canonical
  `ChangeSet` (entries with `EntryRef` subjects, old/new values, positions), the resolved
  `facts` map (each fact typed `resolved|unavailable|invalid|expired` per Spike C, OQ-17),
  `mr` metadata, and the binding's `require:` obligation list — with **no field naming a
  predicate backend** (`cel`, `rego`, or similar) anywhere in the schema, so a backend
  substitution can never require an `EvaluationInput` schema bump.
  - Test: `schemas/decision/v1alpha1/evaluation-input.schema.json`
  - Verify: `go test ./schemas/... -run TestEvaluationInputSchema`
  - Level: L0
- **REQ-P3-E1-S02-02** — Given OQ-8/OQ-9, when
  `schemas/decision/v1alpha1/decision-record.schema.json` is authored, then it requires `decision: enum
  [APPROVE, REVIEW, BLOCK]`, a `findings[]` list (`rule`, `obligation` or none, `effect`,
  `subject` as `EntryRef`, `points`), and a `pins` object requiring `toolVersion`,
  `toolDigest`, `policySha`, `sourceSha`, `targetSha`, `mergeResultDigest` (nullable only when
  the forge capability is absent — ADR-0017 §1), and per-provider `factsResolvedAt`. Redacted
  by construction: no field may carry a raw fact `value` for any fact declared
  `sensitive: true` (cross-reference to the provider envelope's `sensitive` flag, S03).
  Adversarial case: a `decision: APPROVE` record with a null `mergeResultDigest` and no
  companion `capabilityGap` marker fails validation (silent widening is not representable).
  - Test: `schemas/decision/v1alpha1/decision-record.schema.json`
  - Verify: `go test ./schemas/... -run TestDecisionRecordSchema`
  - Level: L0
- **REQ-P3-E1-S02-03** — Given ADR-0016 §3, when `schemas/decision/v1alpha1/replay-bundle.schema.json`,
  `schemas/decision/v1alpha1/presentation-model.schema.json`, and
  `schemas/decision/v1alpha1/publication-receipt.schema.json` are authored, then: `ReplayBundle`
  requires the full unredacted `EvaluationInput` plus every resolved fact value (hermetic
  replay — no wall-clock or environment reference, only the pinned `asOf`/`observedAt`
  values already in-record); `PresentationModel` mirrors `DecisionRecord`'s findings without
  any raw fact value and without rendered markdown (rendering stays a separate concern,
  ADR-0016 §4); `PublicationReceipt` records forge operations performed (`kind: enum [thread,
  approval, merge]`, target IDs, timestamps) and carries no secrets, no raw policy
  expressions, and no user-controlled Markdown.
  - Test: `schemas/decision/v1alpha1/replay-bundle.schema.json`
  - Verify: `go test ./schemas/... -run TestReplayPresentationReceiptSchemas`
  - Level: L0

## P3-E1-S03 — Provider envelope + fact-state/maxAge freeze

- **Goal**: promote the Spike C `FactQuery`/`FactResponse` schemas from `hack/spikes/
  provider/` to the frozen `schemas/provider/v1alpha1/` location unchanged in shape, and
  freeze the per-fact-type `maxAge` default table (OQ-17 residual) as a schema-adjacent
  contract the provider host and lint both read.
- **Operator input**: no.
- **Dependencies**: none (Spike C already proved the shapes; this story is promotion +
  freeze, not new design).
- **Definition of done**: `schemas/provider/v1alpha1/{request,response}.schema.json` exist
  with `$id` unchanged (`https://assent.dev/schemas/provider/v1alpha1/...`); Spike C's own
  tests (`TestContract`, `TestStates`, `TestMinimization`) still pass unmodified against the
  promoted files (only the `//go:embed` path changes); the `maxAge` defaults table is
  committed as a normative doc.

Requirements:

- **REQ-P3-E1-S03-01** — Given Spike C's proven `request.schema.json`/`response.schema.json`,
  when they are promoted to `schemas/provider/v1alpha1/`, then the byte content is unchanged
  (diff shows only the file move) and `hack/spikes/provider/schema.go`'s `//go:embed`
  directives are repointed to the new path, keeping `go test ./hack/spikes/provider/...`
  green with zero test-assertion changes.
  - Test: `schemas/provider/v1alpha1/response.schema.json`
  - Verify: `diff <(git show HEAD:hack/spikes/provider/response.schema.json) schemas/provider/v1alpha1/response.schema.json`
  - Level: L0
- **REQ-P3-E1-S03-02** — Given the four fact states (`resolved|unavailable|invalid|expired`,
  ADR-0017 §6), when the promoted `response.schema.json` is re-verified, then it still
  enforces (via `if/then`) that `state: resolved` requires `value` and `expiresAt`, and any
  other state forbids `value` and requires a `reason` string — distinct machine states, never
  a silently absent key. Adversarial case: a response claiming `resolved` with
  `expiresAt <= asOf` remains the host's responsibility to rewrite to `expired`, but the
  *schema* itself still validates the pre-rewrite `resolved` shape (the rewrite is host logic,
  out of schema scope — documented explicitly so a future implementer does not expect the
  schema to catch staleness).
  - Test: `schemas/provider/v1alpha1/response.schema.json`
  - Verify: `go test ./schemas/... -run TestProviderResponseStates`
  - Level: L0
- **REQ-P3-E1-S03-03** — Given Spike C's proposed `maxAge` table, when it is frozen for P2-E5
  input, then `docs/planning/spikes/spike-c-provider.md`'s table (principal/boolean-auth: 1h;
  registry: 24h; sensitive: 15m; global cap: 24h per ADR-0015 §3) is copied verbatim into
  `docs/planning/predicate-scope.md` or a new `docs/planning/provider-contract.md` as the
  **normative** default (the spike doc stays historical evidence), and states explicitly that
  a provider declaration may only shorten, never lengthen, its type's default.
  - Test: `docs/planning/provider-contract.md`
  - Verify: `grep -qi "15m" docs/planning/provider-contract.md && grep -qi "24h" docs/planning/provider-contract.md`
  - Level: doc

## P3-E1-S04 — Typed `ApprovalEvidence` contract (OQ-23, D-017 B5)

- **Goal**: freeze `ApprovalEvidence` as its own versioned schema — per-rule forge threshold
  (`approvalsRequired`), actual approvers (`approvedBy[]`), principal identity with required
  `isAuthor`, approval source/rule, eligibility evidence, **canonical DecisionRecord pins**
  (cross-file `$ref`), observation time (+ optional time-bound `expiresAt`), and the verifying
  forge capability — so `require-review` (ADR-0017 §3) has one contract no forge adapter can
  bypass with a weaker proxy (like a resolved discussion). Amended 2026-07-22 after roast
  `inbox/reports/2026-07-22-roast-p3-e1-s04-approval-evidence.md` (P1-A/P1-B/P2-C).
- **Operator input**: no.
- **Dependencies**: P3-E1-S02 (`pins` `$def` must exist to `$ref`); GitLab dossier §4
  (`docs/planning/forge-dossier-gitlab.md`, P1-E3-S02).
- **Definition of done**: `schemas/approval/v1alpha1/approval-evidence.schema.json` committed;
  a multi-approval positive fixture (`approvalsRequired ≥ 2`, `approvedBy` length matching)
  built from the dossier §4 chain validates; adversarial fixtures (omitted `isAuthor`, forked
  subset pins, discussion `ruleType`, capability gap with evidence fields) fail as specified.

Requirements:

- **REQ-P3-E1-S04-01** — Given the GitLab dossier §4 evidence chain, when
  `schemas/approval/v1alpha1/approval-evidence.schema.json` is authored, then it requires (when
  `verifyingCapability` ≠ `none`): `principal {id, username, isAuthor: const false}` (required
  `isAuthor` — adapter must positively assert non-authorship; schema cannot verify against an
  in-record author id), `source {rule, ruleType: enum [regular, code_owner, report_approver,
  any_approver]}`, `eligibility {eligibleApproverIds: array, minItems: 1}`,
  `approvalsRequired` (integer ≥ 1 — forge threshold, dossier step (a)), `approvedBy[]`
  (minItems 1, each `{id, username, isAuthor: const false}` — actual approvers, dossier step
  (c), distinct from eligibility), `pins` as a **cross-file `$ref`** to
  `decision-record.schema.json#/$defs/pins` (one shape only — no forked subset),
  `observedAt`, and `verifyingCapability: enum [approval-rules-api, codeowners, none]`.
  `expiresAt` is **optional** (present only for genuine time-bounds; push-staleness uses
  `pins.sourceSha`). No `approved_at` field — `pins.sourceSha` + `observedAt` subsume the
  dossier's approved_at↔head pairing for v1alpha1. Adversarial: missing
  `approvalsRequired`/`approvedBy`, forked subset pins, or omitted `isAuthor` fail validation.
  - Test: `schemas/approval/v1alpha1/approval-evidence.schema.json`
  - Verify: `go test ./schemas/... -run TestApprovalEvidenceSchema`
  - Level: L0
- **REQ-P3-E1-S04-02** — Given ADR-0017 §3's "challenge is acknowledgement, not
  authorization", when the schema is authored, then it has **no field capable of being
  populated from a resolved-discussion/challenge-thread source** (`source.ruleType` is a
  closed enum containing only forge approval-rule/CODEOWNERS kinds) — structurally preventing
  a future implementer from wiring `challenge` resolution into `require-review` evidence.
  Adversarial case: a fixture attempting `source.ruleType: "discussion-resolved"` fails
  schema validation (value not in the enum).
  - Test: `schemas/approval/v1alpha1/approval-evidence.schema.json`
  - Verify: `go test ./schemas/... -run TestApprovalEvidenceExcludesDiscussion`
  - Level: L0
- **REQ-P3-E1-S04-03** — Given "missing capability → never auto-merge, no silent downgrade to
  challenge" (D-017 B5), when `verifyingCapability: "none"` is set, then the schema requires
  `principal`, `source`, `eligibility`, `approvalsRequired`, and `approvedBy` to be
  **absent** (via `if/then`, not merely empty) — so a `DecisionRecord` consuming this
  evidence can only represent the fail-closed outcome. Adversarial case:
  `verifyingCapability: "none"` with `approvedBy` present fails validation.
  - Test: `schemas/approval/v1alpha1/approval-evidence.schema.json`
  - Verify: `go test ./schemas/... -run TestApprovalEvidenceCapabilityGap`
  - Level: L0

## P3-E1-S05 — Adopter test-fixture format schema (ADR-0014)

- **Goal**: freeze the `.assent/tests/` directory-case and inline `cases.yaml` shapes,
  and `expect.yaml`'s obligation/predicate/finding-code assertions with `exact` as the safety
  default (ADR-0017 consequences, replacing the old `vouch`-era `expect.yaml`), as a versioned
  JSON Schema — this is the format `assent test` (E6) and every shipped example pack (E9,
  dogfooding) will be validated against.
- **Operator input**: no.
- **Dependencies**: P3-E1-S01 (findings reference the same `rule`/`effect`/obligation
  vocabulary as `MergePolicy`).
- **Definition of done**: `schemas/testfixture/v1alpha1/test-expectation.schema.json` committed; the
  ADR-0014 `expect.yaml` example (updated to `prove`/`onFailure` vocabulary) and the inline
  `cases.yaml` example both validate; a `message~:` field remains schema-legal but the schema
  description marks it discouraged-for-safety per the ADR-0014 amendment (documentation, not
  a validation rule — wording assertions are a render-layer concern, ADR-0016 §4).

Requirements:

- **REQ-P3-E1-S05-01** — Given ADR-0014's `expect.yaml`, when
  `schemas/testfixture/v1alpha1/test-expectation.schema.json` is authored, then it requires `decision:
  enum [APPROVE, REVIEW, BLOCK]`, an optional `findings[]` (must-contain by default; the case
  sets `exact: true` for a closed list, per the ADR-0017 consequence that `exact` is the
  safety default), an optional `absent[]` (rules that must not fire), and an optional
  `score: {total, threshold}`. Adversarial case: a `findings[]` entry using the retired
  `effect: vouch` fails validation (only `comment`/`challenge`/`block`, or an obligation
  `prove` reference, are legal per S01's frozen vocabulary).
  - Test: `schemas/testfixture/v1alpha1/test-expectation.schema.json`
  - Verify: `go test ./schemas/... -run TestExpectationSchema`
  - Level: L0
- **REQ-P3-E1-S05-02** — Given the inline shorthand, when the schema covers
  `.assent/tests/<pack>/cases.yaml`, then each `cases[]` entry requires `name`, `file`, `base`,
  `head`, and an `expect` object reusing the exact same `expect.yaml` schema (single source of
  truth — the inline and directory forms are not two contracts).
  - Test: `schemas/testfixture/v1alpha1/test-expectation.schema.json`
  - Verify: `go test ./schemas/... -run TestInlineCasesReuseExpectSchema`
  - Level: L0
- **REQ-P3-E1-S05-03** — Given the ADR-0014 amendment (safety vs. presentation assertions),
  when the schema description for `message~:` is written, then it states explicitly that
  `assent test --coverage` counts only structured safety assertions (`decision`, `rule`,
  `effect`, `findings[].path`, `score`) and that rendered-output goldens belong to
  `assent render` fixtures (ADR-0016 §4) — a comment-only, non-validating contract note so
  future implementers do not conflate the two test layers.
  - Test: `schemas/testfixture/v1alpha1/test-expectation.schema.json`
  - Verify: `grep -qi "presentation" schemas/testfixture/v1alpha1/test-expectation.schema.json`
  - Level: doc

## P3-E1-S06 — `assent lint` hard-error list + schema-validation CI job

- **Goal**: enumerate the complete `assent lint` hard-error list (ADR-0010 amendment) as an
  executable-shaped spec, and wire a CI job that validates every committed example/fixture
  against the Phase-3 schemas on every PR — the two are one story because the lint spec is
  meaningless without something to check it against in CI.
- **Operator input**: no.
- **Dependencies**: P3-E1-S01..S05 (the hard-error list references vocabulary and shapes from
  every schema authored so far).
- **Definition of done**: `docs/planning/lint-hard-errors.md` lists every hard error with its
  triggering condition and the ADR/decision that mandates it; a GitHub Actions job (`.github/
  workflows/schemas.yml` or an addition to the existing CI workflow) runs a schema-validation
  step over `schemas/`, `examples/`, and this epic's own fixtures on every PR.

Requirements:

- **REQ-P3-E1-S06-01** — Given ADR-0010's amendment and ADR-0017 §5, when
  `docs/planning/lint-hard-errors.md` is authored, then it lists, at minimum: obligation
  coverage (a binding's `require:` list has no rule that `prove`s it), reserved-class
  violation (a pack reclassifying the built-in `assent-policy` meta-class), fail-open
  restriction (a controlling/authorization fact provider configured `failure: open`),
  tests-per-rule (a rule with zero cases in `assent test --coverage`), unkeyed lists (`entries:
  {mode: list}` without `identity.pointer`), and undeclared predicate-scope fields (an
  `assert`/`cel` leaf referencing a field outside the S01 predicate-scope table) — each row
  cites the ADR/OQ that mandates it.
  - Test: `docs/planning/lint-hard-errors.md`
  - Verify: `grep -qi "obligation coverage" docs/planning/lint-hard-errors.md && grep -qi "unkeyed" docs/planning/lint-hard-errors.md`
  - Level: doc
- **REQ-P3-E1-S06-02** — Given ADR-0015 §1, when the reserved-class hard error is specified,
  then the doc states the adversarial case explicitly: an MR that both touches `.assent/**`
  and carries a pack rule attempting to route the `assent-policy` class to anything other than
  `block`/`challenge` (never `vouch`/obligation-satisfying) is a hard lint error, independent
  of what any rule's predicate evaluates to — self-weakening policy is caught at lint time,
  not decision time.
  - Test: `docs/planning/lint-hard-errors.md`
  - Verify: `grep -qi "assent-policy" docs/planning/lint-hard-errors.md`
  - Level: doc
- **REQ-P3-E1-S06-03** — Given every schema authored in S01–S05, when the CI job is specified,
  then it: (a) compiles every `schemas/**/*.schema.json` as valid JSON Schema draft 2020-12,
  (b) validates every fixture under `examples/**` and this epic's `examples/contracts/**`
  against its matching schema by `apiVersion`/`kind`, and (c) fails the PR — not merely warns —
  on any drift, so example/schema divergence cannot silently merge.
  - Test: `.github/workflows/schemas.yml`
  - Verify: `grep -qi "schema" .github/workflows/schemas.yml`
  - Level: L2

## P3-E1-S07 — The strict §8 exit-gate fixture + named-consumer compatibility fixture (D-016)

- **Goal**: author the one strict, versioned end-to-end fixture ADR-0017 §8 requires — pinned
  target/merge result, a renamed entry with stable identity inside a multi-entry document, two
  independently required obligations, an expired typed fact, a missing required approval, and
  the expected `DecisionRecord` + redacted `PresentationModel` + publication preconditions —
  plus the one sanitized named-consumer compatibility fixture the D-017/B5 freeze gate
  requires. **This is the Phase-3 exit gate**: no engine code (Phase 5, E1–E9) may be started
  before both fixtures validate against the schemas from S01–S05.
- **Operator input**: no (fixtures are generic/invented per D-002 — never copied from the
  reference system; an operator sanitization audit before the Phase-3 freeze review is
  recommended but not blocking, mirroring P1-E1-S02's pattern).
- **Dependencies**: P3-E1-S01 through P3-E1-S06 (every schema and the CI job must exist first —
  this story is the thing they were all built to validate).
- **Definition of done**: `examples/contracts/d016-strict-fixture/` and `examples/contracts/
  named-consumer-compat/` are committed; both validate in the S06 CI job; the D-016 decision
  entry (`docs/decisions/decisions.md`) is updated to record that the gate fixture exists and
  where; `openspec/specs/later-phases.md`'s P3-E1 paragraph gains a "fixture: committed" note.

Requirements:

- **REQ-P3-E1-S07-01** — Given ADR-0017 §8's six required elements, when
  `examples/contracts/d016-strict-fixture/` is authored, then it contains: an
  `EvaluationInput` fixture with (a) `pins.targetSha`/`pins.mergeResultDigest` populated
  (pinned merge result), (b) a multi-entry document where one `EntryRef` is renamed between
  base/head while its `identity.pointer` proves stable identity, (c) exactly two obligations
  independently `require`d by the binding with two distinct proving rules, (d) one typed fact
  in state `expired` (`expiresAt` in the past relative to the fixture's pinned `asOf`), and
  (e) a missing `ApprovalEvidence` for a binding that `require-review`s it; and a companion
  `DecisionRecord` + `PresentationModel` + publication-preconditions fixture recording the
  expected `REVIEW` (or `BLOCK`) outcome consistent with (c)–(e) — all validating against the
  S01–S05 schemas.
  - Test: `examples/contracts/d016-strict-fixture/decision-record.json`
  - Verify: `go test ./schemas/... -run TestD016StrictFixture`
  - Level: L0
- **REQ-P3-E1-S07-02** — Given the missing-approval element, when the fixture's expected
  `DecisionRecord` is built, then its `decision` is never `APPROVE` — a missing required
  `ApprovalEvidence` must produce `REVIEW` or `BLOCK`, and the fixture's
  `PresentationModel` finding for that obligation states the missing evidence without
  fabricating a `capabilityGap` (this MR did not lack the *capability* — it lacked the actual
  approval; the two failure modes stay distinguishable, per S04-03).
  - Test: `examples/contracts/d016-strict-fixture/decision-record.json`
  - Verify: `go test ./schemas/... -run TestD016MissingApprovalNeverApproves`
  - Level: L0
- **REQ-P3-E1-S07-03** — Given D-017/B5's mandatory compatibility fixture, when
  `examples/contracts/named-consumer-compat/` is authored, then it is a single sanitized
  fixture whose `EvaluationInput`/`DecisionRecord` represent, as **structured fields** (never
  inferred from free-text message bodies, labels, or rule names): a rollout `phase`-shaped
  value, a `profile`-shaped value, a comparison-delta-shaped value, an `ApprovalEvidence`
  instance, a publication-marker-shaped correlation value, and a budget-reservation-shaped
  value — every one of these fields validates against its owning schema from this epic or a
  documented placeholder shape pending P3-E4/P3-E5 (those epics own the final field
  definitions; this fixture only proves the *fields exist as structured data*, not free text).
  Adversarial case: a reviewer greps the fixture for banned inference patterns (e.g. a comment
  string containing `"phase: enforce"` as prose rather than a `phase` field) and finds none.
  - Test: `examples/contracts/named-consumer-compat/evaluation-input.json`
  - Verify: `go test ./schemas/... -run TestNamedConsumerCompatFixtureIsStructured`
  - Level: L0
- **REQ-P3-E1-S07-04** — Given D-016, when the gate closes, then
  `docs/decisions/decisions.md` gains a dated entry stating the strict fixture and the
  compatibility fixture exist, validate in CI, and that Phase 5 (E1–E9) may now begin —
  making the "no engine code before this exists" rule auditable rather than a convention
  nobody checks.
  - Test: `docs/decisions/decisions.md`
  - Verify: `grep -qi "D-016" docs/decisions/decisions.md && grep -qi "strict fixture" docs/decisions/decisions.md`
  - Level: doc
