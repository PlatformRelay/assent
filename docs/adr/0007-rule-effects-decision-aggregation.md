# ADR-0007: Rule effects and decision aggregation (incl. risk points)

| | |
| --- | --- |
| **Status** | Proposed |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0002 policy surface](0002-policy-frontends-rego-declarative.md) · [ADR-0008 routing](0008-change-classification-routing-scope.md) · [ADR-0005 forge](0005-forge-abstraction-gitlab-first.md) |

## Context

"A rule fired" is not one thing. In practice a rule's outcome is one of: an informational
comment; a resolvable "are you sure?" thread that must be acknowledged before merge; a hard
block (with optional explanation); or a positive signal that a change is safe to auto-approve.
Severity levels don't capture this — these are **effects** with different forge actions and
different aggregation semantics. Additionally, a scalar **risk score** is wanted so that many
small oddities can add up to "human, please look" even when no single rule blocks.

## Decision (proposed)

### Effects (per rule, declared in the envelope)

| Effect | Meaning | Forge action | Blocks merge? |
| --- | --- | --- | --- |
| `comment` | informational finding | plain MR/PR comment (batched) | no |
| `challenge` | "are you sure?" | **resolvable thread**; merge waits for resolution | until resolved |
| `block` | hard deny | thread/comment (optional custom message) + deny | yes |
| `vouch` | this change is understood and safe | none (recorded in report) | enables automerge |
| `score` | contribute risk points (`points: N`) | none (recorded) | via threshold |

A rule = match + predicate + **one effect** (plus optional `points`, allowed alongside any
effect). Findings carry rule id, effect, paths, message, points.

### Aggregation (deterministic, order-independent)

1. Any `block` finding → **BLOCK**.
2. Else any unresolved `challenge` → **REVIEW** (threads posted; on the forge the MR merges
   only after all threads are resolved *and* re-evaluation passes).
3. Else **coverage check**: every entry in the ChangeSet must be matched by ≥1 `vouch` rule.
   Unvouched changes → **REVIEW** with an explicit "uncovered change" finding. Fail-safe by
   construction: an empty or non-matching policy set never automerges anything.
4. Else **risk check**: `sum(points)` ≤ threshold for the active (environment, change class)
   binding (ADR-0008) → **APPROVE** (+ merge); over threshold → **REVIEW**.

`comment` findings are emitted in every outcome. The JSON report always contains the full
finding list, per-rule traces, score arithmetic, and the aggregation path taken.

## Consequences

- Auto-approve is not an effect a single rule grants unilaterally; it is the *aggregate* of
  full vouch coverage + no blocks/challenges + score under threshold. This keeps the
  default-deny posture while letting packs be composed.
- The forge port (ADR-0005) must support: batched comments, resolvable threads, deny, approve
  + merge — already in scope; `challenge` resolution semantics on GitHub is OQ-7.
- Open sub-questions: score scale conventions and whether thresholds may also *downgrade*
  (e.g. force `challenge` → `block` in prod) — OQ-13.

## Counterpoints considered

- *"Severity levels are simpler."* — They conflate presentation with control flow; mapping
  severities to forge actions per repo would end up re-inventing exactly this effect table,
  but implicitly.
- *"Vouch-coverage is annoying; default-allow with deny rules is less work."* — Default-allow
  automerge on config repos is how outages happen; annoyance is the feature.
