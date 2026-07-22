# P3-E4 — Policy lifecycle contracts: rollout phase, profiles, comparison (ADR-0018)

**Problem**: OQ-21 is reversed by D-017 (B2): rollout is an explicit `off`/`observe`/`enforce`
phase field on rules/packs, not effect-editing — editing effects to simulate rollout loses
policy identity and breaks before/after comparison. The named consumer also needs **named
policy profiles** (exactly one authorizes forge writes; the rest are counterfactual
recorders) and a **semantic policy comparison** (closed delta taxonomy + a versioned,
promotion-gated `PolicyComparisonSuite`) so a pack change can be judged safe *before* it
reaches `enforce`. None of this exists as a contract yet; P3-E1's schemas and ADR-0017's
DecisionRecord/ReplayBundle shapes must grow to carry it.
**Scope**: the openspec stories (INVEST + REQ) that drive — as later implementation work —
the phase field + lint rule, the DecisionRecord observed/enforcing split, the profile +
precedence-table schema, the comparison-delta taxonomy + `PolicyComparisonSuite` format +
promotion gates, ADR-0018 itself, and the named-consumer compatibility-fixture scenarios
this epic owns. **Non-goals**: writing ADR-0018's content now (a later *implementation* lane
authors and files it — this lane only specifies what it must contain, per the operator
brief); the `assent compare` runner or any engine code (Phase 5, E6); editing
`later-phases.md`, other epics' spec dirs, or the P3-E1-owned schema/fixture files
themselves (this epic only adds requirements that reference them).
ADRs: D-017 (B2–B4); OQ-21 (reversed); ADR-0017 §2–4 (obligations, aggregation, one-shot
arming — the enforcing side these contracts must not weaken); ADR-0014 (adopter test
format + its presentation-split amendment); new **ADR-0018** (authored + accepted at the
Phase-3 freeze review by a later lane, per
[named-consumer-compat.md](../../../docs/planning/named-consumer-compat.md) B2–B4).

## P3-E4-S01 — Rollout phase field + lint + DecisionRecord observed/enforcing split

- **Goal**: `off`/`observe`/`enforce` becomes a required, schema-level field on rules and
  packs with exactly the semantics D-017 (B2) fixes — `off` parsed and linted but never
  evaluated; `observe` evaluated and recorded but structurally incapable of altering the
  enforcing decision or forge state; `enforce` contributes obligations, blocks, reviews,
  challenges, and score — and DecisionRecord grows a field-level split between observed and
  enforcing findings so a rule's phase transition is visible in a plain structural diff of
  two loaded-policy snapshots, with no bespoke diff logic required.
- **Operator input**: no.
- **Dependencies**: P3-E1-S01/S02 (authored-resource + DecisionRecord schemas must exist to
  extend); P3-E1-S06 (`docs/planning/lint-hard-errors.md` — S01-04 adds this epic's row to
  that doc, not a separate one); ADR-0017 §2–4 (the enforcing aggregation this must not
  weaken).
- **Definition of done**: the phase field is required (no default — every rule states its
  phase explicitly, consistent with ADR-0017 §9's strict-decode posture) on `MergePolicy`
  rules (extending P3-E1-S01's schema) and on a new `schemas/policy/v1alpha1/pack.schema.json`
  this epic introduces for `pack.yaml` (ADR-0010's pack manifest is not one of P3-E1's three
  authored-resource schemas — P3-E4 is where its schema is first specified, driven by D-017
  (B2)'s "rules/**packs**" wording); `assent lint`'s hard-error list (P3-E1-S06's
  `docs/planning/lint-hard-errors.md`) gains a named `no-implicit-enforce-phase` row; and
  DecisionRecord's finding schema (extending P3-E1-S02) tags every finding with the phase it
  fired under, with only `enforce`-phase findings eligible as inputs to the aggregation
  fields (ADR-0007) — `observe`-phase findings are structurally excluded, not merely excluded
  by convention.

Requirements:

- **REQ-P3-E4-S01-01** — Given a rule declaration, when it is parsed and linted, then the
  schema requires a `phase` field with exactly the enum `off | observe | enforce`, missing or
  unknown values are a hard lint error (never a silent default), and an `off`-phase rule's
  predicate is never invoked by the evaluator (compile/lint-checked only).
  - Test: `schemas/policy/v1alpha1/merge-policy.schema.json`
  - Verify: `grep -q '"phase"' schemas/policy/v1alpha1/merge-policy.schema.json && grep -q '"enum": \["off", "observe", "enforce"\]' schemas/policy/v1alpha1/merge-policy.schema.json`
  - Level: doc
- **REQ-P3-E4-S01-02** — Given a rule at `phase: observe`, when it fires, then its finding is
  recorded in DecisionRecord's `findings.observed` array with the outcome it *would* have had,
  and no field it populates is readable by the aggregation logic that produces `decision`,
  `blocks`, `requiredReviews`, or `score.total` — the adversarial case is a `phase: observe`
  rule declared `effect: block` on every case in the archetype corpus: the resulting
  DecisionRecord's `decision`/`blocks`/`score` are byte-identical to the same corpus with that
  rule deleted entirely.
  - Test: `schemas/decision/v1alpha1/decision-record.schema.json`
  - Verify: `grep -q '"findings"' schemas/decision/v1alpha1/decision-record.schema.json && grep -q '"observed"' schemas/decision/v1alpha1/decision-record.schema.json && grep -q '"enforcing"' schemas/decision/v1alpha1/decision-record.schema.json`
  - Level: doc
- **REQ-P3-E4-S01-03** — Given two loaded-policy snapshots that differ only in one rule's
  `phase` value, when they are diffed with the generic structural differ (no phase-aware
  special case), then the diff surfaces exactly that field change — proving phase transitions
  are visible in policy diffs without a bespoke comparison path, per D-017 (B2)'s "phase
  transitions are visible in policy diffs" requirement.
  - Test: `docs/planning/policy-lifecycle-phase.md`
  - Verify: `grep -qi "phase transition" docs/planning/policy-lifecycle-phase.md && grep -q "structural diff" docs/planning/policy-lifecycle-phase.md`
  - Level: doc
- **REQ-P3-E4-S01-04** — Given `assent lint`'s hard-error list (consolidated per
  `docs/planning/lint-hard-errors.md`, P3-E1-S06), when a pack author edits a rule's
  `effect`/`onFailure` to approximate rollout instead of using `phase`, then no lint rule can
  detect that intent directly (it is a modeling choice, not a syntax error) — so the
  documented contract instead makes the *sanctioned* path strictly easier: the hard-error list
  gains a named rule, `no-implicit-enforce-phase`, that rejects any rule/pack document missing
  an explicit `phase` field (REQ-P3-E4-S01-01/S01-05 already make omission a hard error at the
  schema level; this REQ requires the *lint documentation* to name and cross-reference that
  rule so operators can find why an undecorated rule fails), closing the ambiguity the
  effect-editing anti-pattern exploited.
  - Test: `docs/planning/lint-hard-errors.md`
  - Verify: `test -f docs/planning/lint-hard-errors.md && grep -q "no-implicit-enforce-phase" docs/planning/lint-hard-errors.md`
  - Level: doc
- **REQ-P3-E4-S01-05** — Given a pack manifest (`pack.yaml`, ADR-0010) declaring its own
  `phase` alongside rules that declare theirs, when policy loads, then the pack's `phase`
  acts as a **ceiling**, never additive: a pack at `off` evaluates none of its rules
  regardless of any rule's own phase; a pack at `observe` caps every contained rule at
  `observe` even if the rule itself declares `enforce`; only a pack at `enforce` lets each
  rule's own declared phase stand. Adversarial case: a rule declared `phase: enforce` inside
  a pack declared `phase: observe` must never contribute to `decision`/`blocks`/`score.total`
  — same exclusion guarantee as REQ-P3-E4-S01-02, composed one level up.
  - Test: `schemas/policy/v1alpha1/pack.schema.json`
  - Verify: `grep -q '"phase"' schemas/policy/v1alpha1/pack.schema.json && grep -q '"enum": \["off", "observe", "enforce"\]' schemas/policy/v1alpha1/pack.schema.json`
  - Level: doc

## P3-E4-S02 — Named policy profiles + single-writer rule + precedence table

- **Goal**: a `PolicyProfile` authored resource names a coherent activation of packs/bindings
  distinct from the enforcing policy; exactly one profile may have write authority (forge
  writes: approve, merge, block) for a given (environment, class) binding at any time,
  every other profile is recorder-only (evaluated for comparison, never touches the forge);
  and one precedence table resolves profile × environment/class-binding interaction —
  profiles compose with the existing routing model (ADR-0008) rather than adding a second one.
- **Operator input**: no.
- **Dependencies**: P3-E4-S01 (profiles activate phase-tagged rules); P3-E1 (Config schema).
- **Definition of done**: `PolicyProfile` schema exists with a `writes: boolean` (or
  equivalent authoritative/recorder role) field; lint enforces the single-writer invariant
  (hard error if two profiles both resolve `writes: true` for the same environment/class
  binding); the precedence table is one schema-level artifact on `Config`, not a parallel
  routing mechanism, and its resolution algorithm is documented with a worked example.

Requirements:

- **REQ-P3-E4-S02-01** — Given a `Config` with N declared `PolicyProfile` resources, when
  loaded, then exactly one resolves to `writes: true` for any given (environment, class)
  binding — `assent lint` hard-errors if zero or more than one profile would hold write
  authority for the same binding (adversarial case: two profiles both unconditionally
  `writes: true` with overlapping scope → lint failure, never last-one-wins).
  - Test: `schemas/policy/v1alpha1/profile.schema.json`
  - Verify: `grep -q '"writes"' schemas/policy/v1alpha1/profile.schema.json && grep -q '"required"' schemas/policy/v1alpha1/profile.schema.json`
  - Level: doc
- **REQ-P3-E4-S02-02** — Given a recorder-only (counterfactual) profile, when it is evaluated
  alongside the writing profile over the same ChangeSet, then its DecisionRecord carries a
  `profile` identity field and its outcome is provably inert on the forge — no code path
  reachable from a `writes: false` profile's evaluation calls Reconcile (ADR-0017 §7); this is
  asserted as an architectural invariant, not merely a runtime check.
  - Test: `docs/architecture/policy-profiles.md`
  - Verify: `grep -qi "recorder-only" docs/architecture/policy-profiles.md && grep -qi "Reconcile" docs/architecture/policy-profiles.md`
  - Level: doc
- **REQ-P3-E4-S02-03** — Given a repo with profiles scoped to different (environment, class)
  bindings, when precedence is resolved, then one documented table (`profile ×
  environment/class-binding`) — not a second `match`/routing block — determines the winner,
  and the worked example shows at least one case where a narrower-scoped profile wins over a
  broader one plus one case where scopes are disjoint (both profiles active, at most one with
  `writes: true`).
  - Test: `docs/planning/policy-lifecycle-profiles.md`
  - Verify: `grep -qi "precedence table" docs/planning/policy-lifecycle-profiles.md && grep -qi "worked example" docs/planning/policy-lifecycle-profiles.md`
  - Level: doc

## P3-E4-S03 — Comparison delta taxonomy + PolicyComparisonSuite format + promotion gates

- **Goal**: a semantic `ComparisonRecord` contract with a **closed** delta taxonomy —
  `stricter-intervention-added`, `destructive-or-authorization-intervention-missed`,
  `subject-or-obligation-uncovered`, `newly-auto-mergeable`, `score-threshold-change`,
  `explanation-only` — classifies every difference between a baseline and a candidate
  profile's decisions over the same corpus; a versioned `PolicyComparisonSuite` format pins an
  immutable corpus of `ReplayBundle`s (P3-E1) with stable case IDs; and machine-enforceable
  promotion gates (zero missed destructive interventions, zero missed
  authorization/ownership interventions, no unexpected obligation removal, bounded auto-merge
  widening, explicitly accepted deltas) are data the suite carries, not prose a human
  interprets. Rendered-message changes are structurally excluded from ever producing a
  non-`explanation-only` delta (ADR-0014's amendment split).
- **Operator input**: no.
- **Dependencies**: P3-E4-S01/S02 (a comparison is baseline profile vs candidate profile, both
  phase-aware); P3-E1 (ReplayBundle schema).
- **Definition of done**: `ComparisonRecord` schema exists with the six-member closed enum
  (`additionalProperties`/`enum` — no open string field, per ADR-0017 §9's do-not-generalize
  posture) and per-delta identity (which rule/obligation/subject, baseline vs candidate
  outcome); `PolicyComparisonSuite` schema exists with stable `caseId`s and an immutable-corpus
  invariant (adding a case never mutates an existing `caseId`'s ReplayBundle); the promotion
  gates are expressed as a schema-level pass/fail table keyed by delta kind, consumed later
  by `assent compare` (E6) — never by prose alone.

Requirements:

- **REQ-P3-E4-S03-01** — Given a baseline and a candidate decision over the same ChangeSet,
  when they differ, then the difference is classified into exactly one of the six closed
  taxonomy members — `assent compare`'s future implementation has no "other"/free-text delta
  kind, and a difference matching none of the six is itself a hard error (fail-closed
  classification, never silently dropped).
  - Test: `schemas/comparison/v1alpha1/comparison-record.schema.json`
  - Verify: `grep -q '"enum"' schemas/comparison/v1alpha1/comparison-record.schema.json && grep -q "destructive-or-authorization-intervention-missed" schemas/comparison/v1alpha1/comparison-record.schema.json && grep -q "explanation-only" schemas/comparison/v1alpha1/comparison-record.schema.json`
  - Level: doc
- **REQ-P3-E4-S03-02** — Given a wording-only change to a rule's `message` template (no
  predicate/effect/phase/obligation change), when compared, then the delta classifies as
  `explanation-only` and no other taxonomy member — proving rendered-message changes never
  enter the semantic gates, per the ADR-0014 amendment split.
  - Test: `schemas/comparison/v1alpha1/comparison-record.schema.json`
  - Verify: `grep -qi "message" schemas/comparison/v1alpha1/comparison-record.schema.json`
  - Level: doc
- **REQ-P3-E4-S03-03** — Given a `PolicyComparisonSuite` document, when a new case is added in
  a later revision, then every prior `caseId`'s `ReplayBundle` content hash is unchanged
  (immutable corpus) — the adversarial case is a suite revision that edits an existing
  `caseId`'s fixture in place: this must be represented in the format as a new `caseId`, never
  an in-place edit, so historical promotion-gate results stay reproducible.
  - Test: `schemas/comparison/v1alpha1/comparison-suite.schema.json`
  - Verify: `grep -q '"caseId"' schemas/comparison/v1alpha1/comparison-suite.schema.json && grep -qi "immutable" schemas/comparison/v1alpha1/comparison-suite.schema.json`
  - Level: doc
- **REQ-P3-E4-S03-04** — Given the five machine-enforceable promotion gates (zero missed
  destructive; zero missed authorization/ownership; no unexpected obligation removal; bounded
  auto-merge widening; explicitly accepted deltas), when a suite run classifies its deltas,
  then each gate maps to a deterministic pass/fail function of delta kind + explicit
  acceptance list — `destructive-or-authorization-intervention-missed` deltas always fail
  their gate unless individually present in an `acceptedDeltas` allowlist keyed by case ID and
  delta identity (never by kind alone, closing the "accept all deltas of this kind" footgun).
  - Test: `docs/planning/policy-lifecycle-promotion-gates.md`
  - Verify: `grep -qi "promotion gate" docs/planning/policy-lifecycle-promotion-gates.md && grep -qi "acceptedDeltas" docs/planning/policy-lifecycle-promotion-gates.md`
  - Level: doc

## P3-E4-S04 — ADR-0018 authorship + freeze-acceptance criteria + `assent compare` CLI spec

- **Goal**: schedule, as a later *implementation* lane's Definition of Done, the drafting and
  filing of **ADR-0018** (phase/profile/comparison contract) plus the freeze-acceptance
  criteria that gate its move to Accepted at the Phase-3 freeze review, and the doc-level
  `assent compare` CLI spec (inputs: baseline + candidate profile refs, a
  `PolicyComparisonSuite` ref; outputs: a comparison report; exit codes tied 1:1 to the
  promotion-gate outcomes of P3-E4-S03) whose *implementation* is Phase 5+ (E6) but whose
  *contract* freezes now, per the epic paragraph's story seeds. This story does not create
  `docs/adr/0018-*.md` — see Non-goals.
- **Operator input**: no (ADR-0018's *acceptance* is an operator gate at the Phase-3 freeze
  review, same as P2-E5 for ADR-0002–0017 — that is a later story, not this one).
- **Dependencies**: P3-E4-S01/S02/S03 (ADR-0018's decision sections are exactly those three
  contracts); named-consumer-compat.md B2–B4 (the disposition ADR-0018 must record).
- **Definition of done**: this story's own DoD is the *spec*, not the ADR: a later
  implementation lane can draft `docs/adr/0018-policy-lifecycle-phase-profile-comparison.md`
  straight from these REQs with no further design work, following the [ADR
  template](../../../docs/adr/template.md) and the ADR index conventions in
  [docs/adr/README.md](../../../docs/adr/README.md); that same lane drafts the CLI spec
  section; freeze-acceptance happens in a distinct later story (the P2-E5 pattern: author
  first, operator ratifies at the gate).

Requirements:

- **REQ-P3-E4-S04-01** — Given P3-E4-S01/S02/S03's contracts, when ADR-0018 is drafted by a
  later lane, then it exists at `docs/adr/0018-policy-lifecycle-phase-profile-comparison.md`,
  follows `docs/adr/template.md`'s section structure, is added to the `docs/adr/README.md`
  index table with status `Proposed`, and its Decision section addresses all three of: the
  phase field + DecisionRecord split, profiles + single-writer + precedence, and the
  comparison delta taxonomy + `PolicyComparisonSuite` + promotion gates — no partial ADR.
  - Test: `docs/adr/0018-policy-lifecycle-phase-profile-comparison.md`
  - Verify: `test -f docs/adr/0018-policy-lifecycle-phase-profile-comparison.md && grep -q "0018" docs/adr/README.md`
  - Level: doc
- **REQ-P3-E4-S04-02** — Given ADR-0018 in `Proposed` status, when the Phase-3 freeze review
  runs (the later acceptance story, mirroring P2-E5-S02), then its status line flips to
  `Accepted` (or a stated partial-supersession note) only after the schemas from S01–S03 exist
  and validate, and named-consumer-compat.md's B2–B4 disposition rows are updated from "new
  P3-E4" links to the accepted ADR-0018 link.
  - Test: `docs/adr/README.md`
  - Verify: `grep -q "0018" docs/adr/README.md`
  - Level: doc
- **REQ-P3-E4-S04-03** — Given the frozen comparison contract (S03), when the `assent compare`
  CLI spec is authored (doc-level; implementation is E6/Phase 5+), then it names: the two
  required profile-ref inputs, the `PolicyComparisonSuite` ref input, a comparison-report
  output shape reusing `ComparisonRecord`, and an exit-code table with one code per
  promotion-gate failure kind plus a distinct all-pass code — `assent compare` runs are
  declared side-effect-free (no Reconcile calls, ever) in the same spec.
  - Test: `docs/adr/0018-policy-lifecycle-phase-profile-comparison.md`
  - Verify: `grep -qi "assent compare" docs/adr/0018-policy-lifecycle-phase-profile-comparison.md && grep -qi "side-effect-free" docs/adr/0018-policy-lifecycle-phase-profile-comparison.md`
  - Level: doc

## P3-E4-S05 — Compatibility-fixture scenarios: observe-vs-enforce + refused auto-merge-widening

- **Goal**: the one sanitized named-consumer compatibility fixture in the mandatory Phase-3
  fixture set — owned and authored by P3-E1 at `examples/contracts/named-consumer-compat/`
  (its `evaluation-input.json`/`decision-record.json` pair, per P3-E1-S07-03) — is extended,
  by requirements this story states rather than by this lane editing P3-E1's files, to
  exercise two P3-E4-specific scenarios: (a) the same ChangeSet evaluated once at
  `phase: observe` and once at `phase: enforce` producing identical `decision`/forge-state
  but different `findings.observed` content, and (b) a candidate profile that would widen
  auto-merge eligibility (a `newly-auto-mergeable` delta) being refused by the promotion
  gates because the widening is not in the case's `acceptedDeltas` allowlist.
- **Operator input**: no.
- **Dependencies**: **P3-E1-S07** (owns `examples/contracts/named-consumer-compat/` and its
  exact file names; this story only adds requirements those already-planned files must
  satisfy — no edits to P3-E1's paths from this lane); P3-E4-S01 (phase semantics + the
  `findings.observed`/`findings.enforcing` split), P3-E4-S03 (delta taxonomy + promotion
  gates + `acceptedDeltas`).
- **Definition of done**: `examples/contracts/named-consumer-compat/decision-record.json`
  (P3-E1-owned file) carries at least one case pair demonstrating (a) and one case
  demonstrating (b), using exactly the field names S01/S03 define
  (`findings.observed`/`findings.enforcing`, `newly-auto-mergeable`, `acceptedDeltas`); its
  sibling `evaluation-input.json` carries `phase` and `profile` as typed fields; neither file
  infers phase, profile, or comparison-delta information from message text, labels, or rule
  names — they are structured fields, per D-017's fixture requirement and P3-E1-S07-03's own
  adversarial check.

Requirements:

- **REQ-P3-E4-S05-01** — Given the named-consumer compatibility fixture's
  `decision-record.json`, when it represents an `observe`-phase case and its `enforce`-phase
  twin (same ChangeSet, same rule, only `phase` differs), then `decision`, `blocks`,
  `requiredReviews`, and `score.total` are byte-identical between the two, while
  `findings.observed` is populated only in the `observe` case and the corresponding entry
  appears in `findings.enforcing` only in the `enforce` case — the adversarial check is that
  swapping which case is "observe" and which is "enforce" without swapping the phase field
  must fail this comparison, proving the assertion is not vacuously true.
  - Test: `examples/contracts/named-consumer-compat/decision-record.json`
  - Verify: `test -f examples/contracts/named-consumer-compat/decision-record.json && grep -q '"observed"' examples/contracts/named-consumer-compat/decision-record.json && grep -q '"enforcing"' examples/contracts/named-consumer-compat/decision-record.json`
  - Level: doc
- **REQ-P3-E4-S05-02** — Given a candidate profile whose comparison against the baseline
  yields a `newly-auto-mergeable` delta for one case represented in the fixture, when that
  case's `acceptedDeltas` allowlist does not name it, then the promotion-gate evaluation for
  that case fails (bounded auto-merge widening gate, S03) and
  `examples/contracts/named-consumer-compat/decision-record.json` records the
  refused-widening outcome explicitly — never a silent pass because "newly auto-mergeable"
  sounds like an improvement.
  - Test: `examples/contracts/named-consumer-compat/decision-record.json`
  - Verify: `test -f examples/contracts/named-consumer-compat/decision-record.json && grep -q "newly-auto-mergeable" examples/contracts/named-consumer-compat/decision-record.json && grep -q "acceptedDeltas" examples/contracts/named-consumer-compat/decision-record.json`
  - Level: doc
- **REQ-P3-E4-S05-03** — Given both fixture scenarios, when inspected for D-017's
  structured-field requirement (the same requirement P3-E1-S07-03 already states for approval
  evidence and marker fields), then `phase` and `profile` are typed fields in
  `examples/contracts/named-consumer-compat/evaluation-input.json` — never derivable only
  from a rule name, a label, or message text; a reviewer grepping the fixture pair for
  prose-inference patterns (e.g. a message string containing `"phase: enforce"` instead of a
  `phase` field) finds none, mirroring P3-E1-S07-03's own adversarial check.
  - Test: `examples/contracts/named-consumer-compat/evaluation-input.json`
  - Verify: `test -f examples/contracts/named-consumer-compat/evaluation-input.json && grep -q '"phase"' examples/contracts/named-consumer-compat/evaluation-input.json && grep -q '"profile"' examples/contracts/named-consumer-compat/evaluation-input.json`
  - Level: doc
