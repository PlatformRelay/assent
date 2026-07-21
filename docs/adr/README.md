# Architecture Decision Records

Format: see [`template.md`](template.md). Numbered, immutable once **Accepted** (supersede,
don't edit). Statuses: `Proposed` → `Accepted` → (`Superseded by ADR-nnnn`).

Most ADRs start **Proposed** during the design phase and are firmed up via trade-off matrices
in the planning workflow (see [`../planning/meta-plan.md`](../planning/meta-plan.md)).

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-language-go-single-binary.md) | Implementation language: Go, single static binary | Accepted |
| [0002](0002-policy-frontends-rego-declarative.md) | Policy frontends: Rego + Kyverno-style declarative YAML over one engine | Proposed |
| [0003](0003-canonical-change-model.md) | Canonical change model for JSON / YAML / HCL-tfvars | Proposed |
| [0004](0004-plugin-architecture.md) | Plugin architecture for permission & fact providers | Proposed |
| [0005](0005-forge-abstraction-gitlab-first.md) | Forge abstraction: GitLab first, GitHub second | Proposed |
| [0006](0006-testing-strategy.md) | Testing strategy: spec-driven pyramid with real-forge e2e | Proposed |
