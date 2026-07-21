# P2-E5 — ADR acceptance round (0002–0017)

**Problem**: the Phase-2 gate: every Proposed ADR moves to Accepted or Superseded with
matrix + spike evidence; amendments and supersessions (ADR-0017 reshapes seven earlier
ADRs) must be reconciled so no two accepted ADRs contradict.
**Scope**: evidence consolidation + operator review + status flips + OQ/decision-log
hygiene. **Non-goals**: new design work (anything unresolved becomes a Phase-3 spec item or
a new Proposed ADR, not a blocker edit).
ADRs: all; D-011, D-016.

## P2-E5-S01 — Consolidate evidence and reconcile supersessions

- **Goal**: one review document mapping every Proposed ADR (0002–0017) to its evidence
  (spike REQs, dossier sections, prior-art lessons), its open counterpoints, and — where
  ADR-0017 reshapes it — the exact clauses superseded (e.g. ADR-0007 §3 vouch coverage,
  ADR-0012 override mechanism) so acceptance can mark partial supersessions precisely.
- **Operator input**: no.
- **Dependencies**: P2-E1..E4 complete; P1-E3, P1-E4.
- **Definition of done**: `docs/planning/adr-acceptance-review.md` has one row per ADR
  0002–0017 with evidence links and a recommended verdict; every OQ tagged "ADR-xxxx accept"
  has a proposed resolution in the doc.

Requirements:

- **REQ-P2-E5-S01-01** — Given the spike outputs, when the review doc is authored, then it
  contains a table row for each of ADR-0002 through ADR-0017 with columns `Evidence`,
  `Superseded clauses`, `Recommended verdict`.
  - Test: `docs/planning/adr-acceptance-review.md`
  - Verify: `grep -q "ADR-0017" docs/planning/adr-acceptance-review.md && grep -q "Recommended verdict" docs/planning/adr-acceptance-review.md`
  - Level: doc

## P2-E5-S02 — Operator review: flip statuses, close OQs (the Phase-2 gate)

- **Goal**: operator reviews the consolidated evidence; each ADR's status line becomes
  `Accepted` or `Superseded by ADR-nnnn`; the ADR index table, open-questions table, and
  decision log are updated in the same change.
- **Operator input**: **yes** — verdicts are the operator's.
- **Dependencies**: P2-E5-S01.
- **Definition of done**: no ADR remains `Proposed`; resolved OQs struck through with
  resolution links; a dated D-nnn entry records the acceptance round; residual questions
  re-tagged to Phase-3 gates.

Requirements:

- **REQ-P2-E5-S02-01** — Given the review outcome, when statuses are flipped, then
  `docs/adr/README.md` contains no `Proposed` status.
  - Test: `docs/adr/README.md`
  - Verify: `! grep -q "Proposed" docs/adr/README.md`
  - Level: doc
- **REQ-P2-E5-S02-02** — Given the acceptance round, when the decision log is updated, then
  `docs/decisions/decisions.md` contains a dated entry referencing the ADR acceptance round
  and the spike evidence.
  - Test: `docs/decisions/decisions.md`
  - Verify: `grep -qi "acceptance round" docs/decisions/decisions.md`
  - Level: doc
