# P3-E5 — Publication reconciliation protocol (database-free)

**Problem**: D-017 (B6) commits to freezing the marker + reconciliation protocol that the
Reconcile port (ADR-0017 §7) and the finding-lifecycle state machine (ADR-0012, ADR-0011
`UpsertComment`/`SyncThreads`) already assume, while preserving D-007 (no database — the
forge itself, via hidden-HTML markers on bot-authored comments, is the durable reconciliation
surface). Without a frozen grammar and a numbered contract, "rerun idempotence" (the P4-E1
exit gate) has nothing precise to implement against, and duplicate-comment incidents have no
deterministic repair rule.
**Scope**: the marker grammar (four concepts), the reconciliation state table + numbered
protocol as schema + fixtures, the duplicate-repair rule + one-publisher-per-MR topology +
`doctor` guarantee reporting, and the authorship criteria for **ADR-0019** (the ADR itself,
the walkthrough update, and the doctor-checklist row are later impl-lane Definition-of-Done
items — this epic specifies what they must contain).
**Non-goals**: authoring ADR-0019 itself; implementing the Reconcile port, `assent doctor`,
or `serve`'s keyed lock (E4/E12, Phase 5); the general schema-authoring/validation-CI-job
machinery for all contracts (P3-E1 owns that tooling — this epic freezes only the marker and
reconciliation *content*); GitHub marker parity (E10, **Locked** D-012 — the dossier
(P1-E3-S03) already keeps that seam honest).
ADRs: 0011 (`UpsertComment`/`SyncThreads` lifecycle), 0012 (finding-key state machine, marker
comments, secret redaction), 0015 §6 (serve dedup key, idempotent publishing), 0016 §1
(renderer-owned marker region), 0017 §7 (`Reconcile(DesiredReviewState, Preconditions) ->
PublicationReceipt`); D-007, D-017 (B6), D-018.

## P3-E5-S01 — Marker grammar: four concepts, non-goals, spoofing ignore

- **Goal**: freeze the hidden-HTML marker grammar around four distinct, independently
  identifiable concepts — `slot` (stable identity from canonical fields: project/MR, rule ID,
  obligation, EntryRef, effect, anchor), `occurrence` (hash of the safety-relevant occurrence,
  so changed content cannot inherit a prior resolution), `decision` (the DecisionRecord hash
  that requested this state), `artifact` (kind + schema version) — and state, as load-bearing
  non-goals, that markers are correlation metadata only (never decision input or authorization
  evidence), that only bot-authored comments are parsed (contributor marker spoofing is
  ignored), and that markers carry no secrets, fact values, user-controlled Markdown, or raw
  policy expressions.
- **Operator input**: no.
- **Dependencies**: none (freezes concepts already implied by ADR-0011/0012/0016).
- **Definition of done**: `docs/contracts/p3-e5-publication-protocol/marker-grammar.md` names
  each of the four concepts with its exact field composition and a worked example marker
  payload; `docs/contracts/p3-e5-publication-protocol/marker-grammar.schema.json` is valid
  JSON Schema encoding that payload shape; both the occurrence-supersession and the
  spoofing-ignored behaviours are pinned by adversarial prose + example, not left implicit.

Requirements:

- **REQ-P3-E5-S01-01** — Given the four marker concepts, when the grammar is authored, then
  `marker-grammar.md` documents each of `slot`/`occurrence`/`decision`/`artifact` with its
  exact source fields (slot: project/MR, rule ID, obligation, EntryRef, effect, anchor;
  occurrence: hash of the safety-relevant occurrence content; decision: DecisionRecord hash;
  artifact: kind + schema version) and `marker-grammar.schema.json` is a syntactically valid
  JSON Schema whose required properties are exactly those four concepts.
  - Test: `docs/contracts/p3-e5-publication-protocol/marker-grammar.md`,
    `docs/contracts/p3-e5-publication-protocol/marker-grammar.schema.json`
  - Verify: `grep -q "slot" docs/contracts/p3-e5-publication-protocol/marker-grammar.md && grep -q "occurrence" docs/contracts/p3-e5-publication-protocol/marker-grammar.md && python3 -c "import json; s=json.load(open('docs/contracts/p3-e5-publication-protocol/marker-grammar.schema.json')); req=set(s['required']); assert req == {'slot','occurrence','decision','artifact'}, req"`
  - Level: doc
- **REQ-P3-E5-S01-02** — Given a slot whose entry content changes between runs (e.g. an
  obligation's target value is edited after a `challenge` was posted and resolved), when the
  occurrence hash is recomputed, then it differs from the prior occurrence and the grammar
  doc's adversarial example states explicitly that the new occurrence **cannot inherit** the
  old occurrence's resolved state — the prior resolution stays scoped to the superseded
  occurrence only.
  - Test: `docs/contracts/p3-e5-publication-protocol/marker-grammar.md`
  - Verify: `grep -qi "cannot inherit" docs/contracts/p3-e5-publication-protocol/marker-grammar.md`
  - Level: doc
- **REQ-P3-E5-S01-03** — Given the marker's role in reconciliation, when the non-goals are
  documented, then `marker-grammar.md` states in one explicit section that markers are
  **correlation metadata only** — matching a forge artifact to the slot/occurrence/decision
  that produced it — and are **never** decision input or authorization evidence; a rule
  predicate, an obligation proof, or `require-review` eligibility must never read a marker
  value as if it were a fact, an approval, or a DecisionRecord field.
  - Test: `docs/contracts/p3-e5-publication-protocol/marker-grammar.md`
  - Verify: `grep -qi "never.*decision input" docs/contracts/p3-e5-publication-protocol/marker-grammar.md || grep -qi "correlation metadata only" docs/contracts/p3-e5-publication-protocol/marker-grammar.md`
  - Level: doc
- **REQ-P3-E5-S01-04** — Given a contributor (non-bot identity) posts a comment containing a
  well-formed marker for an existing slot (e.g. claiming a BLOCK-backing challenge is already
  resolved), when reconciliation lists bot-authored artifacts, then the contributor's comment
  is excluded by construction (author-identity filter, not marker content) and has zero effect
  on the computed reconciliation state — the adversarial fixture makes this the explicit
  expected outcome, not an assumed side effect.
  - Test: `docs/contracts/p3-e5-publication-protocol/marker-grammar.md`
  - Verify: `grep -qi "spoof" docs/contracts/p3-e5-publication-protocol/marker-grammar.md && grep -qi "author" docs/contracts/p3-e5-publication-protocol/marker-grammar.md`
  - Level: doc
- **REQ-P3-E5-S01-05** — Given comment bodies are visible to any project member and subject to
  CI artifact retention, when prohibited marker content is documented, then `marker-grammar.md`
  lists secrets, fact values, user-controlled Markdown, and raw policy expressions as content
  that must never appear in a marker payload, and `marker-grammar.schema.json` structurally
  cannot represent free-form/arbitrary string fields wide enough to carry them (every property
  is a bounded enum, hash, or ID string).
  - Test: `docs/contracts/p3-e5-publication-protocol/marker-grammar.md`,
    `docs/contracts/p3-e5-publication-protocol/marker-grammar.schema.json`
  - Verify: `grep -qi "secrets" docs/contracts/p3-e5-publication-protocol/marker-grammar.md && grep -qi "raw policy expression" docs/contracts/p3-e5-publication-protocol/marker-grammar.md`
  - Level: doc

## P3-E5-S02 — Reconciliation state table + numbered protocol fixtures

- **Goal**: freeze the reconciliation semantics as a numbered contract (recompute
  DesiredReviewState from trusted inputs; list paginated bot-authored artifacts; update the
  one summary slot in place; leave the same unresolved occurrence untouched; preserve
  resolution of the same occurrence across reruns; supersede stale occurrences with fresh
  challenges; resolve no-longer-desired findings; deterministically repair pre-existing
  duplicates — REQs in S03; rescan after publication before reporting success) and a state
  table of (existing-artifact-state x desired-state) -> action, each cell backed by a fixture.
  This is the exit-gate artifact P4-E1 consumes for its rerun-idempotence case.
- **Operator input**: no.
- **Dependencies**: P3-E5-S01 (occurrence/slot vocabulary).
- **Definition of done**: `docs/contracts/p3-e5-publication-protocol/reconciliation-state-table.md`
  lists all nine numbered steps in order with the invariant each protects; a state table
  covers every (artifact-state x desired-state) combination named in the epic text; two
  fixtures under `docs/contracts/p3-e5-publication-protocol/fixtures/` — rerun-idempotence and
  crash-then-rerun — each give a before/after artifact list with zero duplicate creation.

Requirements:

- **REQ-P3-E5-S02-01** — Given the epic's protocol text, when the numbered contract is
  authored, then `reconciliation-state-table.md` lists exactly nine numbered steps in the
  stated order (recompute DesiredReviewState; list paginated bot-authored artifacts; update
  one summary slot in place; leave same unresolved occurrence untouched; preserve resolution
  across reruns; supersede stale occurrences; resolve no-longer-desired findings;
  deterministically repair pre-existing duplicates; rescan after publication before success),
  and each step names the invariant it protects (no-database/D-007, determinism, or
  target-ref trust).
  - Test: `docs/contracts/p3-e5-publication-protocol/reconciliation-state-table.md`
  - Verify: `test "$(grep -c '^[0-9]\+\.' docs/contracts/p3-e5-publication-protocol/reconciliation-state-table.md)" = "9"`
  - Level: doc
- **REQ-P3-E5-S02-02** — Given an existing bot-authored artifact and a freshly computed
  DesiredReviewState, when the state table enumerates their combinations (no-artifact /
  artifact-matches-current-occurrence-unresolved / artifact-matches-current-occurrence-resolved
  / artifact-has-stale-occurrence / artifact-no-longer-desired), then each row names exactly
  one action (create / leave-untouched / preserve-resolution / supersede-with-fresh-challenge /
  resolve) and the doc states that no row triggers more than one action.
  - Test: `docs/contracts/p3-e5-publication-protocol/reconciliation-state-table.md`
  - Verify: `grep -qi "no-longer-desired" docs/contracts/p3-e5-publication-protocol/reconciliation-state-table.md && grep -qi "stale occurrence" docs/contracts/p3-e5-publication-protocol/reconciliation-state-table.md`
  - Level: doc
- **REQ-P3-E5-S02-03** — Given a DesiredReviewState computed identically twice in a row (a
  plain rerun with no source change), when the rerun-idempotence fixture is authored, then
  `fixtures/rerun-idempotence.yaml` records run 1's created artifacts and states run 2's
  expected output as zero new artifacts and zero duplicate slot occupancy — the same fixture
  shape P4-E1's exit gate cites for "a rerun ... produce[s] zero duplicate comments/threads".
  - Test: `docs/contracts/p3-e5-publication-protocol/fixtures/rerun-idempotence.yaml`
  - Verify: `test -f docs/contracts/p3-e5-publication-protocol/fixtures/rerun-idempotence.yaml && grep -qi "zero" docs/contracts/p3-e5-publication-protocol/fixtures/rerun-idempotence.yaml`
  - Level: doc
- **REQ-P3-E5-S02-04** — Given a run that crashes after step 3 (creating some artifacts) but
  before step 9 (the post-publication rescan), when the crash-then-rerun fixture is authored,
  then `fixtures/crash-then-rerun.yaml` shows run 2 first listing existing bot-authored
  artifacts (step 2) and therefore creating nothing new for slots already covered by run 1's
  partial work — matching P4-E1's "crash-then-rerun produce[s] zero duplicate comments/threads".
  - Test: `docs/contracts/p3-e5-publication-protocol/fixtures/crash-then-rerun.yaml`
  - Verify: `test -f docs/contracts/p3-e5-publication-protocol/fixtures/crash-then-rerun.yaml && grep -qi "crash" docs/contracts/p3-e5-publication-protocol/fixtures/crash-then-rerun.yaml`
  - Level: doc

## P3-E5-S03 — Duplicate-repair fixture + one-publisher topology + doctor guarantee reporting

- **Goal**: freeze the deterministic duplicate-repair rule (lowest forge ID canonical; repair
  recorded in PublicationReceipt), the supported one-publisher-per-MR topology (CI
  `resource_group` keyed per MR IID; `serve` mode: a keyed per-MR lock), and the requirement
  that `doctor` reports which duplicate-prevention guarantee the deployment actually provides
  — explicitly naming multi-replica HA as unsupported rather than silently assumed safe.
- **Operator input**: no.
- **Dependencies**: P3-E5-S02 (state table defines "no-longer-desired"/"stale" so "duplicate"
  has a precise meaning: two-or-more bot artifacts occupying the same slot).
- **Definition of done**: `docs/contracts/p3-e5-publication-protocol/fixtures/duplicate-repair.yaml`
  exists with a seeded multi-artifact-per-slot scenario and its canonical-repair outcome;
  `reconciliation-state-table.md` (or a sibling section) states the one-publisher topology and
  the doctor-report field; both are written so a later impl lane can lift them verbatim into
  ADR-0019, the walkthrough, and the doctor checklist (S04).

Requirements:

- **REQ-P3-E5-S03-01** — Given two or more pre-existing bot-authored artifacts occupying the
  same slot (e.g. seeded by a race or a prior bug), when the duplicate-repair fixture is
  authored, then `fixtures/duplicate-repair.yaml` names the lowest-forge-ID artifact as
  canonical, every other duplicate as repaired (resolved/removed), and includes an expected
  `PublicationReceipt.repairs` list recording each repaired artifact's forge ID and the
  canonical ID it was resolved against — repair is deterministic, not first-seen-wins by scan
  order.
  - Test: `docs/contracts/p3-e5-publication-protocol/fixtures/duplicate-repair.yaml`
  - Verify: `test -f docs/contracts/p3-e5-publication-protocol/fixtures/duplicate-repair.yaml && grep -qi "lowest" docs/contracts/p3-e5-publication-protocol/fixtures/duplicate-repair.yaml && grep -qi "repairs" docs/contracts/p3-e5-publication-protocol/fixtures/duplicate-repair.yaml`
  - Level: doc
- **REQ-P3-E5-S03-02** — Given strict duplicate prevention requires exactly one publisher per
  MR, when the topology is documented, then `reconciliation-state-table.md` states the CI
  `resource_group` (keyed per MR IID) mechanism for one-shot mode and the keyed-per-MR-lock
  mechanism for `serve` mode, and explicitly states that multi-replica HA is **unsupported**:
  duplicates created under concurrent unserialized publishers converge only on the next
  reconciliation, never immediately.
  - Test: `docs/contracts/p3-e5-publication-protocol/reconciliation-state-table.md`
  - Verify: `grep -qi "resource_group" docs/contracts/p3-e5-publication-protocol/reconciliation-state-table.md && grep -qi "unsupported" docs/contracts/p3-e5-publication-protocol/reconciliation-state-table.md`
  - Level: doc
- **REQ-P3-E5-S03-03** — Given a deployment's actual serialization cannot always be verified
  from inside the process (e.g. `resource_group` misconfigured, or `serve` running without its
  keyed lock), when the doctor-guarantee requirement is documented, then
  `reconciliation-state-table.md` specifies a typed report field (e.g.
  `duplicate_prevention: single-writer-serialized | unserialized-best-effort`) that `doctor`
  must emit, with the adversarial case stated: doctor must never claim
  `single-writer-serialized` when it cannot verify the serialization mechanism is actually
  configured — the safe default on ambiguity is `unserialized-best-effort`.
  - Test: `docs/contracts/p3-e5-publication-protocol/reconciliation-state-table.md`
  - Verify: `grep -qi "duplicate_prevention" docs/contracts/p3-e5-publication-protocol/reconciliation-state-table.md && grep -qi "unserialized-best-effort" docs/contracts/p3-e5-publication-protocol/reconciliation-state-table.md`
  - Level: doc

## P3-E5-S04 — ADR-0019 authorship criteria + walkthrough/doctor serialization requirements

- **Goal**: specify — without writing it — what a later impl lane's ADR-0019 must contain to
  be accepted at the Phase-3 freeze review (per `named-consumer-compat.md`: "New ADRs are
  authored inside their owning epics ... accepted at the Phase-3 freeze review"), and require
  that the setup walkthrough and the `doctor` checklist fold in the serialization/duplicate-
  prevention requirement frozen in S03, so an adopter cannot reach a false sense of safety.
- **Operator input**: no (criteria only; ADR ratification itself is the Phase-3 freeze review,
  same governance as ADR-0002-0017 acceptance).
- **Dependencies**: P3-E5-S01, S02, S03 (the ADR restates their frozen content; it does not
  design new content).
- **Definition of done**: this spec states the acceptance criteria precisely enough that a
  reviewer can check ADR-0019, the walkthrough, and the doctor checklist against it without
  re-deriving the protocol; none of those three target files are modified by this lane.

Requirements:

- **REQ-P3-E5-S04-01** — Given S01-S03's frozen content, when ADR-0019 is authored by a later
  lane, then `docs/adr/0019-publication-marker-reconciliation-protocol.md` follows
  `docs/adr/template.md`'s structure, cites D-007 and D-017 (B6) in Context links, and its
  Decision section states the four-concept marker split, the nine-step reconciliation
  contract, and the one-publisher-per-MR topology as three separately numbered decisions (not
  folded into one paragraph) so each can be superseded independently later.
  - Test: `docs/adr/0019-publication-marker-reconciliation-protocol.md`
  - Verify: `grep -q "D-007" docs/adr/0019-publication-marker-reconciliation-protocol.md && grep -qi "one-publisher" docs/adr/0019-publication-marker-reconciliation-protocol.md`
  - Level: doc
- **REQ-P3-E5-S04-02** — Given ADR-0019 is proposed inside the Phase-3 freeze review scope
  (not the P2-E5 acceptance round, which covers only ADR-0002-0017), when the ADR index is
  updated, then `docs/adr/README.md` gains a row for ADR-0019 whose Status matches the ADR's
  own status line (`Proposed` until the freeze review, then `Accepted`) — the two files must
  never disagree.
  - Test: `docs/adr/README.md`
  - Verify: `grep -q "0019" docs/adr/README.md`
  - Level: doc
- **REQ-P3-E5-S04-03** — Given the one-publisher-per-MR requirement (S03), when the CI wiring
  step of the setup walkthrough is updated, then `docs/usage/walkthrough.md`'s CI step states
  the `resource_group` (or `serve` keyed-lock) requirement explicitly, not merely implied by
  the surrounding YAML — an adopter copying the walkthrough without reading ADR-0019 must
  still end up serialized.
  - Test: `docs/usage/walkthrough.md`
  - Verify: `grep -qi "resource_group" docs/usage/walkthrough.md`
  - Level: doc
- **REQ-P3-E5-S04-04** — Given the doctor-guarantee report field (S03), when the doctor
  checklist is updated, then `docs/planning/spikes/spike-secure-setup.md`'s `## Doctor
  checklist` table gains a precondition row for `duplicate_prevention` with its evidence
  source (e.g. `resource_group` config on the CI job, or the `serve` lock backend) and failure
  consequence (`warn` when unverifiable, never a silent `single-writer-serialized` claim).
  - Test: `docs/planning/spikes/spike-secure-setup.md`
  - Verify: `grep -qi "duplicate_prevention" docs/planning/spikes/spike-secure-setup.md`
  - Level: doc
