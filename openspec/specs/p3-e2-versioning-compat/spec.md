# P3-E2 — Versioning & compatibility spec

**Problem**: ADR-0017 §9 commits to compatibility rules — safety-bearing resources reject
unknown fields/duplicate IDs/unknown enums; reports are additive-tolerant; provider majors are
negotiated; named collections are lists with mandatory unique IDs; hashes are over canonical
JSON with schema-version domain separation; a do-not-generalize list is frozen — but leaves
them as prose. Without executable fixtures per rule, "compat rule" is unfalsifiable and P3-E1's
schemas could drift from ADR-0017 §9 unnoticed; adopters also have no public document stating
what a version bump may and may not break.
**Scope**: fixture suites (valid + adversarial-invalid) proving each ADR-0017 §9 compat rule;
canonical-JSON hash vectors with schema-version domain separation; a provider major-version
negotiation matrix; the public `API_STABILITY.md` (oss-playbook #8); the do-not-generalize list
restated as a named lint-error catalog.
**Non-goals**: authoring the JSON Schemas themselves (P3-E1); the production loader/lint code
path that enforces these rules at runtime (Phase 5, E3) — this epic specifies the executable
*fixture contract* those consumers must satisfy, not the loader implementation; example/pack
migration (P3-E3); rollout-phase/profile/comparison contracts (P3-E4); publication
reconciliation (P3-E5).
ADRs: ADR-0017 §6 (typed provider protocol — majors), §7 (contract ownership — schemas are the
public API, no v1 SDK), §9 (compatibility rules & scope guards); D-016; oss-playbook #8.

## P3-E2-S01 — Strict-decode fixture suite for safety-bearing resources

- **Goal**: prove, with executable valid + adversarial-invalid fixtures, that every
  safety-bearing authored resource (Config, RulesetBinding, MergePolicy — the P3-E1 schema set)
  rejects unknown fields, duplicate IDs, and unknown enum values at decode time — never
  silently dropped or coerced.
- **Operator input**: no.
- **Dependencies**: P3-E1 (the schemas fixtures decode against).
- **Definition of done**: `schemas/testdata/compat/strict-decode/` has ≥1 valid + ≥1
  adversarial-invalid fixture per rule (unknown field, duplicate ID, unknown enum) per
  safety-bearing resource type; the suite fails if any adversarial fixture decodes cleanly, and
  fails if any valid fixture is rejected.

  > **Path amend (D-016, 2026-07-22)**: D-016/ADR-0017 §8 forbid engine code under
  > `internal/core` before the §8 exit-gate fixture (P3-E1-S07) lands, so this story's fixtures
  > and tests were relocated to the `schemas` package (contract-tooling, not engine code) — see
  > `agent-context/INBOX.md` "P3-E2 fixture tests stay under schemas/ (D-016)". Fixtures live at
  > `schemas/testdata/compat/strict-decode/`; tests at `schemas/compat_strictdecode_test.go`,
  > reusing the existing `ConfigSchema`/`RulesetBindingSchema`/`MergePolicySchema` validators.
  > Revert to `internal/core/schema/...` only after the Phase-3 gate closes.

Requirements:

- **REQ-P3-E2-S01-01** — Given a MergePolicy/RulesetBinding/Config fixture carrying one field
  absent from its P3-E1 schema, when the strict-decode suite runs, then decoding is rejected
  with a positioned error naming the unknown field.
  - Test: `schemas/compat_strictdecode_test.go`
  - Verify: `go test ./schemas/... -run TestStrictDecode/unknown_field`
  - Level: L0
- **REQ-P3-E2-S01-02** — Given a fixture whose named collection (entries/rules/obligations)
  carries two elements sharing one ID, when decoded, then the suite rejects it as a
  duplicate-ID violation. Adversarial case: the duplicate differs only in an unrelated field, so
  a naive last-write-wins merge would otherwise hide it.
  - Test: `schemas/compat_strictdecode_test.go`
  - Verify: `go test ./schemas/... -run TestStrictDecode/duplicate_id`
  - Level: L0
- **REQ-P3-E2-S01-03** — Given a fixture using an enum value not declared by its resource's
  schema (e.g. an `effect` or `onFailure.effect` outside the frozen set), when decoded, then the
  suite rejects it rather than coercing to a default or nearest-known value.
  - Test: `schemas/compat_strictdecode_test.go`
  - Verify: `go test ./schemas/... -run TestStrictDecode/unknown_enum`
  - Level: L0
- **REQ-P3-E2-S01-04** — Given the same three rules, when a fixture with no unknown
  field/duplicate ID/unknown enum is decoded, then it succeeds unchanged — the suite asserts
  both directions so a rule cannot be satisfied by a validator that rejects everything.
  - Test: `schemas/compat_strictdecode_test.go`
  - Verify: `go test ./schemas/... -run TestStrictDecode/valid`
  - Level: L0

## P3-E2-S02 — Additive-tolerant reports + unique-ID named collections

- **Goal**: prove reports (DecisionRecord, ReplayBundle, PresentationModel, PublicationReceipt,
  the generated rule catalogue) decode successfully when they carry fields unknown to an older
  consumer schema — additive/forward tolerance — while named collections, in both authored and
  report schemas, remain lists with a mandatory unique ID field, rejected at decode/lint when
  unkeyed or duplicated; source order never carries implicit identity.
- **Operator input**: no.
- **Dependencies**: P3-E1 (report schemas); P3-E2-S01 (shared strict-decode harness for the
  collection-identity half).
- **Definition of done**: fixture pairs exist for each report kind and each
  named-collection-bearing schema; ≥1 additive-tolerant valid fixture (unknown top-level field)
  decodes; ≥1 unkeyed-list and ≥1 duplicate-ID adversarial fixture per collection is rejected.

Requirements:

- **REQ-P3-E2-S02-01** — Given a DecisionRecord/ReplayBundle/PresentationModel/
  PublicationReceipt fixture carrying an extra top-level field unknown to a schema-version-N
  consumer, when decoded by that consumer, then decoding succeeds and the unknown field is
  preserved rather than causing a hard failure — proving reports are additive-tolerant while
  authored resources (S01) are not.
  - Test: `internal/core/schema/reporttolerance_test.go`
  - Verify: `go test ./internal/core/schema/... -run TestReportAdditiveTolerant`
  - Level: L0
- **REQ-P3-E2-S02-02** — Given a named collection (e.g. `entries`, `findings`, `obligations`)
  declared as a bare unkeyed array, when validated, then it is rejected at decode/lint.
  Adversarial case: reordering the array must never change which element a later reference
  resolves to, because the rule forbids the unkeyed shape entirely rather than tolerating it.
  - Test: `internal/core/schema/collectionidentity_test.go`
  - Verify: `go test ./internal/core/schema/... -run TestUniqueIDCollections/unkeyed`
  - Level: L0
- **REQ-P3-E2-S02-03** — Given a named collection with two elements sharing an ID and no
  explicit `priority` field, when validated, then it is rejected — source order carries no
  meaning absent an explicit tiebreaker.
  - Test: `internal/core/schema/collectionidentity_test.go`
  - Verify: `go test ./internal/core/schema/... -run TestUniqueIDCollections/duplicate`
  - Level: L0

## P3-E2-S03 — Canonical-JSON hash vectors + schema-version domain separation

- **Goal**: fix one canonicalization algorithm (key ordering, whitespace, number formatting)
  and prove hashes are computed over canonical JSON with the schema version folded into the
  hash domain, via a checked-in vector table — identical logical content re-serialized
  differently hashes identically; the same content under two different schema versions never
  collides.
- **Operator input**: no.
- **Dependencies**: P3-E1 (the schema-version field the domain separation keys on).
- **Definition of done**: `internal/core/hash/vectors.json` has ≥6 cases (reordered keys,
  whitespace variants, numeric-formatting variants, two schema versions of the same logical
  document, one deliberately non-canonical input) with an expected hash per case; the hasher
  matches every row in CI.

Requirements:

- **REQ-P3-E2-S03-01** — Given two JSON encodings of the same logical document differing only
  in key order and whitespace, when each is canonicalized and hashed, then both yield the
  identical hash recorded in the vector table.
  - Test: `internal/core/hash/vectors_test.go`
  - Verify: `go test ./internal/core/hash/... -run TestCanonicalHash/stable`
  - Level: L0
- **REQ-P3-E2-S03-02** — Given the same logical document serialized once under schema-version
  `v1alpha1` and once under a bumped version string with no content change, when hashed, then
  the two recorded hashes differ. Adversarial case: a naive hash-over-bytes-only implementation
  would collide here, hiding a schema-version downgrade/replay.
  - Test: `internal/core/hash/vectors_test.go`
  - Verify: `go test ./internal/core/hash/... -run TestCanonicalHash/domain_separation`
  - Level: L0
- **REQ-P3-E2-S03-03** — Given the vector table, when `internal/core/hash` changes, then CI
  fails any row whose recomputed hash no longer matches the checked-in expected value — hash
  regressions are caught, never silently re-baselined.
  - Test: `internal/core/hash/vectors.json`
  - Verify: `go test ./internal/core/hash/... -run TestCanonicalHash`
  - Level: L0

## P3-E2-S04 — Provider major-version negotiation matrix

- **Goal**: specify, as an executable matrix, how a provider announcing protocol major version
  `P` negotiates against a host built for major version `H` — every `(H, P)` cell resolves to
  exactly one of `accept` (`P == H`) or `capability gap → provider refused, facts it would have
  supplied become unavailable, never auto-merge on them` (`P != H`); no cross-major coercion, no
  partial acceptance.
- **Operator input**: no.
- **Dependencies**: none new (consumes the typed provider envelope, ADR-0017 §6, from P2-E3 /
  P3-E1).
- **Definition of done**: negotiation fixture matrix covers the host's current major plus at
  least one older and one newer hypothetical major, plus a missing/unparseable-major case.

Requirements:

- **REQ-P3-E2-S04-01** — Given a provider response declaring the host's current major version,
  when negotiated, then the host accepts and processes facts normally.
  - Test: `internal/provider/negotiation_test.go`
  - Verify: `go test ./internal/provider/... -run TestMajorNegotiation/match`
  - Level: L0
- **REQ-P3-E2-S04-02** — Given a provider response declaring a different major version (older
  or newer) than the host, when negotiated, then the host reports a capability gap for that
  provider and treats every fact it would have supplied as `unavailable` — never silently
  reinterpreted under the host's schema.
  - Test: `internal/provider/negotiation_test.go`
  - Verify: `go test ./internal/provider/... -run TestMajorNegotiation/mismatch`
  - Level: L0
- **REQ-P3-E2-S04-03** — Given a provider response with a missing or unparseable major-version
  field, when negotiated, then the host treats it identically to a mismatch (fail closed).
  Adversarial case: a provider that omits the field to dodge the check must not be rewarded
  with acceptance.
  - Test: `internal/provider/negotiation_test.go`
  - Verify: `go test ./internal/provider/... -run TestMajorNegotiation/missing`
  - Level: L0

## P3-E2-S05 — `API_STABILITY.md` + do-not-generalize lint-error catalog

- **Goal**: publish the public `API_STABILITY.md` (oss-playbook #8) stating per-version
  guarantees for the policy schema, decision contract, and adopter test format, plus graduation
  criteria (e.g. draft → frozen leaving `v1alpha1`); restate ADR-0017 §9's do-not-generalize
  list (no user-defined effects, custom aggregators, obligation `anyOf`, generic
  entry-selector query language, LCD forge API) as named, stable `assent lint` hard-error codes
  so proposing any of them is caught by lint, not only by a design review that might be skipped.
- **Operator input**: no (drafted here; ratified alongside the Phase-3 freeze review per the
  meta-plan).
- **Dependencies**: P3-E2-S01..S04 (the guarantees this doc states are exactly the rules those
  suites enforce).
- **Definition of done**: `API_STABILITY.md` committed at repo root with the three guarantee
  tables, a graduation-criteria section, and a "Do-not-generalize (lint errors)" section with
  one stable error code per do-not-generalize item, cross-linked to ADR-0017 §9.

Requirements:

- **REQ-P3-E2-S05-01** — Given the frozen compat rules (S01–S04), when `API_STABILITY.md` is
  authored, then it contains one table row per public contract (policy schema, decision
  contract, adopter test format) stating the current version, the compatibility guarantee
  (e.g. "additive-only within a major", "strict-decode, no unannounced breaking change"), and
  graduation criteria for leaving `v1alpha1`.
  - Test: `API_STABILITY.md`
  - Verify: `grep -q "policy schema" API_STABILITY.md && grep -qi "graduation" API_STABILITY.md`
  - Level: doc
- **REQ-P3-E2-S05-02** — Given ADR-0017 §9's do-not-generalize list, when the lint-error
  catalog section is authored, then each of the five items (user-defined effects, custom
  aggregators, obligation `anyOf`, generic entry-selector query language, LCD forge API) has a
  named stable error code and a one-line rationale referencing ADR-0017 §9.
  - Test: `API_STABILITY.md`
  - Verify: `grep -qi "do-not-generalize" API_STABILITY.md && grep -q "ADR-0017" API_STABILITY.md`
  - Level: doc
- **REQ-P3-E2-S05-03** — Given a future change that adds one of the five do-not-generalize
  items to `assent lint`'s accepted config surface, when the exit-gate fixture suite (S01)
  runs against a fixture exercising that surface, then the fixture is present as a permanent
  adversarial-invalid case per catalog entry — removing the guard shows up as a fixture diff,
  never a silent regression.
  - Test: `schema/fixtures/compat/do-not-generalize/`
  - Verify: `go test ./internal/core/schema/... -run TestDoNotGeneralize`
  - Level: L0
