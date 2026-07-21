# Examples

Living documentation. Rules are written in the **YAML envelope** (ADR-0002 v2/ADR-0013):
`assert` predicates (CEL leaves, string shorthand) for the archetypes, `rego` modules where
logic outgrows tier 1 — the rego example here shows that escape hatch. Every example must
pass the adopter test harness (`assent test`) once it exists — examples that don't run are
lies.

> ⚠️ Pre-alpha: schemas are **illustrative drafts** anchored by ADR-0010/0013; the
> authoritative contracts are frozen in meta-plan Phase 3.

- [`policies/declarative/`](policies/declarative/) — envelope rules with `assert` predicates
- [`policies/rego/`](policies/rego/) — the tier-2 escape hatch for the same archetype
- [`repos/`](repos/) — generic sample self-service repo layouts (generated; e2e seeds)
