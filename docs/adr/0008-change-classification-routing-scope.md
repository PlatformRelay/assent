# ADR-0008: Change classification, ruleset routing, and rule scope

| | |
| --- | --- |
| **Status** | Proposed |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0003 change model](0003-canonical-change-model.md) · [ADR-0007 effects](0007-rule-effects-decision-aggregation.md) |

## Context

One repo holds many kinds of entries; one MR may touch several. "This part of the diff is a
Kafka-topic change" must route those changes into the topic ruleset; a tfvars edit under
`prod/` must route into a stricter binding than the same edit under `dev/`. Separately, some
rules must see more than the diff: a naming-convention rule should be able to comment on a
touched entry whose *pre-existing* state violates convention, which requires the full branch
state, not just the changed fields.

## Decision (proposed)

### 1. Classification stage

After the differ (ADR-0003), a **classifier** assigns each ChangeSet entry one or more
**change classes** via declarative matchers (path globs + content predicates, e.g.
`file: topics/**` or `has(new.partitions)`), and detects the **environment** (path convention
or repo config). Classes and environments are just labels — repos define their own.
Unclassified changes get the implicit class `unclassified` (which no vouch rule should match
→ fail-safe REVIEW per ADR-0007).

### 2. Ruleset routing

A **`RulesetBinding`** document maps `(change class, environment)` → policy packs + risk
threshold. One MR touching topics *and* tfvars evaluates both packs, each over its slice of
the ChangeSet; aggregation (ADR-0007) runs over the union of findings with the **strictest
matching threshold**.

### 3. Rule scope

Each rule declares `scope`:

- `change` (default) — predicate sees the matched ChangeSet entries (old/new values).
- `branch` — predicate additionally sees the **full repo state at the head SHA** (parsed
  value trees of the checked-out branch). For: conventions on touched-but-not-changed fields,
  cross-entry uniqueness, referential checks inside the repo.

Determinism holds: head SHA is pinned input; branch scope is still a pure function.

### 4. Local checkout is mandatory

Evaluation always runs against a **local checkout of the MR source branch** (merged-result
checkout where the forge supports it). CI has this for free; webhook mode (ADR-0009) clones
per event. No API-only file fetching — partial views breed nondeterminism and n+1 API calls.

## Consequences

- The classifier is the routing seam: packs stay small and per-domain (topic pack, catalog
  pack, tfvars pack) and can be shared/versioned independently — the marketplace unit.
- `branch`-scoped rules are costlier (parse the tree); the engine parses lazily per class/glob.
- Cross-*repo* state stays out of scope: that is what fact providers are for (ADR-0004).

## Counterpoints considered

- *"Put match globs on every rule instead of a classifier stage."* — Works at small scale,
  but environment × class × pack routing then lives half-duplicated inside every rule;
  bindings centralize it and make the risk-threshold table explicit and auditable.
