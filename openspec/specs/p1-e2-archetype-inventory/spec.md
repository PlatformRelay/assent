# P1-E2 — Rule-archetype inventory + success metric

**Problem**: the archetype inventory is the acceptance bar for the whole policy surface
(meta-plan 1.2) and the Phase-1 gate. ADR-0017 changed what an archetype must specify:
each maps to a **named obligation** it proves (`prove`/`onFailure`), a matcher **domain**
(`files` / `values.pointers` / `fileEvents` / `valueChanges`), and a subject (`EntryRef`).
OQ-25 (success metric) folds in here.
**Scope**: inventory doc, per-archetype example change + expected decision, metric definition.
**Non-goals**: JSON Schemas (Phase 3), executable fixtures in the frozen ADR-0014 format
(fixtures here carry DRAFT markers per ADR-0017 consequences).
ADRs: 0007 (as reshaped by 0017 §2–3), 0013 + appendix gallery, 0017 §2/§3/§5, OQ-13, OQ-25.

## P1-E2-S01 — Author the archetype inventory

- **Goal**: enumerate every rule class the vision archetypes + operator shapes + corpus
  require; for each: inputs (change paths, facts, permissions), decision semantics, failure
  mode, obligation mapping, matcher domain, effect + `onFailure`, tri-state error behaviour.
- **Operator input**: no to start (seed from vision.md + corpus); updated when P1-E1-S01
  shapes arrive.
- **Dependencies**: P1-E1-S02 (corpus informs coverage); soft on P1-E1-S01.
- **Definition of done**: inventory covers at minimum ownership, bounded-change,
  allow-listed-fields, no-destruction (delete **and** rename, per ADR-0010 amendment),
  environment-split, schema-validity, freshness/context-fact, and the `assent-policy`
  meta-class (ADR-0015 §1); each names the obligation it proves and whether failure needs
  `require-review` (authorization) vs `challenge` (acknowledgement) per ADR-0017 §3.

Requirements:

- **REQ-P1-E2-S01-01** — Given vision.md's archetype list and the corpus, when the inventory
  is authored, then `docs/planning/archetypes.md` contains one section per archetype with
  the fields: `Inputs`, `Subject (EntryRef)`, `Matcher domain`, `Obligation proved`,
  `onFailure effect+code`, `Failure mode (tri-state)`, `Forge outcome`.
  - Test: `docs/planning/archetypes.md`
  - Verify: `grep -q "Obligation proved" docs/planning/archetypes.md && grep -q "Matcher domain" docs/planning/archetypes.md`
  - Level: doc
- **REQ-P1-E2-S01-02** — Given ADR-0017 §3, when an archetype requires eligible-identity
  authorization (e.g. ownership failure, destructive change), then its section states
  `require-review` — not `challenge` — and cites the forge evidence needed (link to
  P1-E3-S02). Adversarial case documented per such archetype: the author self-resolving a
  thread must not satisfy it.
  - Test: `docs/planning/archetypes.md`
  - Verify: `grep -q "require-review" docs/planning/archetypes.md`
  - Level: doc

## P1-E2-S02 — Example change + expected decision per archetype (the Phase-1 gate)

- **Goal**: every archetype gets ≥1 concrete base/head file pair from a sample-repo shape
  plus the expected decision, findings, and obligation outcomes — DRAFT format, later
  migrated to the frozen fixture format (ADR-0014 / Phase 3).
- **Operator input**: no.
- **Dependencies**: P1-E2-S01, P1-E1 (shapes to draw files from).
- **Definition of done**: one directory per archetype under `examples/archetypes/`; each has
  `base/`, `head/`, `facts.yaml`, `expected.yaml` (with `# DRAFT — pre-schema` header);
  at least one *negative* twin (expected REVIEW/BLOCK) per archetype; the adversarial
  near-threshold rename pair (ADR-0003 amendment) exists for no-destruction.

Requirements:

- **REQ-P1-E2-S02-01** — Given an archetype in the inventory, when its example is authored,
  then `examples/archetypes/<id>/` contains `base/`, `head/`, `facts.yaml`, and
  `expected.yaml` naming decision, firing rule intent, and obligation satisfied/failed.
  - Test: `examples/archetypes/`
  - Verify: `test -d examples/archetypes && find examples/archetypes -name expected.yaml | grep -q expected`
  - Level: doc
- **REQ-P1-E2-S02-02** — Given the no-destruction archetype, when examples are authored,
  then a delete case, a rename case, and an adversarial near-similarity delete+add pair each
  exist with expected decision REVIEW or BLOCK (never APPROVE) — pinning ADR-0003's
  "rename never laxer than delete".
  - Test: `examples/archetypes/no-destruction/expected.yaml`
  - Verify: `test -f examples/archetypes/no-destruction/expected.yaml`
  - Level: doc

## P1-E2-S03 — Success metric definition (OQ-25)

- **Goal**: define the "≥70% of routine MRs" north star measurably: who defines the
  "routine" denominator, the adjudicated holdout protocol, and the false-auto-merge budget
  (ADR-0017 §9).
- **Operator input**: **yes** — operator adjudicates the holdout set labels.
- **Dependencies**: P1-E1-S02 (corpus supplies historical MRs/PRs to sample).
- **Definition of done**: `docs/planning/success-metric.md` defines: denominator rule
  (independent of assent's own classifier), holdout sampling + adjudication protocol
  (labeler ≠ policy author), false-auto-merge budget with response action, and how `scan`'s
  confusion matrix (ADR-0009 amendment 2) reports against it; OQ-25 closed with link.

Requirements:

- **REQ-P1-E2-S03-01** — Given the corpus history, when the metric is defined, then
  `docs/planning/success-metric.md` contains sections `Denominator`, `Holdout protocol`,
  `False-auto-merge budget`, and `Measurement via scan/stats`.
  - Test: `docs/planning/success-metric.md`
  - Verify: `grep -q "Denominator" docs/planning/success-metric.md && grep -qi "budget" docs/planning/success-metric.md`
  - Level: doc
