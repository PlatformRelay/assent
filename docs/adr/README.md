# Architecture Decision Records

Format: see [`template.md`](template.md). Numbered, immutable once **Accepted** (supersede,
don't edit). Statuses: `Proposed` → `Accepted` → (`Superseded by ADR-nnnn`).

Most ADRs start **Proposed** during the design phase and are firmed up via trade-off matrices
and spikes in the planning workflow (see [`../planning/meta-plan.md`](../planning/meta-plan.md)).

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-language-go-single-binary.md) | Implementation language: Go, single static binary | Accepted |
| [0002](0002-policy-frontends-rego-declarative.md) | Policy surface: one Kyverno-style YAML envelope, pluggable expression backends | Proposed (v2) |
| [0003](0003-canonical-change-model.md) | Canonical change model for JSON / YAML / HCL-tfvars (incl. deletions & renames) | Proposed |
| [0004](0004-plugin-architecture.md) | Plugin architecture for permission & fact providers | Proposed |
| [0005](0005-forge-abstraction-gitlab-first.md) | Forge abstraction: GitLab first, GitHub second | Proposed |
| [0006](0006-testing-strategy.md) | Testing strategy: spec-driven pyramid with real-forge e2e | Proposed |
| [0007](0007-rule-effects-decision-aggregation.md) | Rule effects and decision aggregation (incl. risk points) | Proposed |
| [0008](0008-change-classification-routing-scope.md) | Change classification, ruleset routing, and rule scope | Proposed |
| [0009](0009-execution-modes.md) | Execution modes: CI, local/dry-run, explain, webhook, scan/stats | Proposed |
| [0010](0010-config-files-repo-layout.md) | Configuration files and governed-repo layout | Proposed |
| [0011](0011-core-ports-and-contracts.md) | Core Go ports and public contracts (draft shapes) | Proposed |
| [0012](0012-presentation-templates-debug.md) | Presentation: comment rendering, expandable details, docs links, rule debug | Proposed |
| [0013](0013-assert-syntax-and-backend.md) | `assert` syntax and backend: CEL-leaf condition trees on cel-go ([gallery](0013-appendix-syntax-gallery.md)) | Proposed |
| [0014](0014-adopter-test-format.md) | Adopter test format — policy tests as a public contract | Proposed |
| [0015](0015-trust-boundaries-merge-integrity.md) | Trust boundaries and merge-time integrity (from the 2026-07-21 adversarial review) | Proposed |
| [0016](0016-presentation-theming.md) | Presentation theming: config knobs, slots, CEL messages, render contract | Proposed |
| [0017](0017-contract-model-obligations.md) | Contract model: governed subjects, required obligations, typed facts, preconditioned reconciliation | Proposed |
