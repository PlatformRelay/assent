# P1-E4 — Prior-art review

**Problem**: meta-plan 1.4 — steal shamelessly, differentiate deliberately. Each tool's
lessons must land as design implications tied to our ADRs, not as a book report.
**Scope**: one document, per-tool sections + a consolidated implications table.
**Non-goals**: re-litigating decided ADRs (implications may *feed* P2-E5, not reopen D-006).

## P1-E4-S01 — Author `docs/planning/prior-art.md`

- **Goal**: for OPA/conftest, Kyverno (incl. ValidatingPolicy direction), Mergify,
  Bors/merge queues, Renovate automerge, Prow/Tide, danger.js, and the conda-forge/Homebrew
  bot-automerge precedent: what each got right/wrong *for this use case*.
- **Operator input**: no.
- **Dependencies**: none.
- **Definition of done**: every tool section answers: change-model approach, policy surface,
  merge-integrity approach (TOCTOU handling!), testing story for rules, adoption friction;
  implications table maps each lesson to an ADR or open question.

Requirements:

- **REQ-P1-E4-S01-01** — Given the tool list, when the review is authored, then
  `docs/planning/prior-art.md` has one section per tool (all eight present) and each section
  ends with `Lessons for assent:`.
  - Test: `docs/planning/prior-art.md`
  - Verify: `grep -c "Lessons for assent" docs/planning/prior-art.md | grep -qE '^[8-9]|^[1-9][0-9]'`
  - Level: doc
- **REQ-P1-E4-S01-02** — Given the lessons, when the implications table is written, then
  each row cites an ADR number or OQ id it confirms, amends, or challenges — at minimum
  covering merge-queue prior art vs ADR-0017 §1 and Prow/Tide's re-test-before-merge vs
  ADR-0009's arming model.
  - Test: `docs/planning/prior-art.md`
  - Verify: `grep -q "ADR-0017" docs/planning/prior-art.md && grep -qi "tide" docs/planning/prior-art.md`
  - Level: doc
