# Examples

Living documentation. Each rule archetype appears in **both** policy syntaxes where
expressible ([ADR-0002](../docs/adr/0002-policy-frontends-rego-declarative.md)) and every
example must pass the adopter test harness (`assent test`) once it exists — examples that
don't run are lies.

> ⚠️ Pre-alpha: schemas below are **illustrative drafts** to anchor the design discussion.
> The authoritative PolicyInput/rule schemas are frozen in meta-plan Phase 3.

- [`policies/rego/`](policies/rego/) — Rego rules against the draft PolicyInput document
- [`policies/declarative/`](policies/declarative/) — the same rules, Kyverno-style YAML
- [`repos/`](repos/) — generic sample self-service repo layouts (generated; used as e2e seeds)
