# Decision log

Lightweight, dated operator decisions (D-nnn). ADRs cover architecture; this file covers
project/process decisions.

| ID | Date | Decision |
| --- | --- | --- |
| D-001 | 2026-07-21 | Started under working title "verdict-2". **No public GitHub repo until explicitly created by the operator.** Local git only for now. Name resolved by D-009. |
| D-002 | 2026-07-21 | License **Apache-2.0** (Kubernetes/Argo CD precedent: permissive + patent grant; supersedes the initial MIT choice same day). Fully open source; no employer names, internal system names, or internal policy content in any committed artifact. |
| D-003 | 2026-07-21 | Go module path `github.com/PlatformRelay/assent`; confirm org (PlatformRelay vs personal) before first push. |
| D-004 | 2026-07-21 | Git identity for this repo: `Konrad Heimel <konrad.heimel@gmail.com>` (local git config). No AI co-author trailers. |
| D-005 | 2026-07-21 | Development is **spec/test-driven**: openspec workflow, REQ IDs with `Test:` + `Verify:`, TDD for all implementation work. |
| D-006 | 2026-07-21 | Operator preference: **Kyverno-style declarative YAML is the primary policy surface**; Rego is an escape-hatch predicate backend inside the same envelope, never a parallel frontend (ADR-0002 v2). |
| D-007 | 2026-07-21 | **No database for now**: the per-run JSON report artifact is the storage format; `stats` aggregates report files (ADR-0009/0012). |
| D-008 | 2026-07-21 | Sample corpus: operator provides 2–3 further real self-service repo shapes (generalized, never verbatim); additionally curate **open-source repos** as public test/demo corpora (candidates in [examples/repos](../../examples/repos/README.md)). |
| D-009 | 2026-07-21 | Project name: **assent** (chosen from the [naming candidates](../planning/naming.md)). Folder, module path, CLI (`assent`), config dir (`.assent/`), and `apiVersion` group (`assent.dev/v1alpha1`) renamed accordingly. Before going public: verify GitHub name availability and the `assent.dev` domain (fallback: adjust the apiVersion group). |
| D-010 | 2026-07-21 | **Testing bar from day one**: TDD mandatory; CI-enforced **≥90% line coverage** on `internal/…` from the first implementation PR; golden decision tests + integration tests + real-forge e2e per ADR-0006. The **adopter test format is a first-class public contract** (ADR-0014) — easy policy testability for governed repos is a product feature, not tooling. |
| D-011 | 2026-07-21 | **Architecture review cycle completed**: adversarial review (15 findings, 2 critical) folded back as ADR-0015 (trust boundaries) + amendments to ADR-0003/0004/0007/0009/0011/0012; `assert` decided as hybrid CEL-leaf trees on cel-go (ADR-0013, kyverno-json dropped); engineering rules codified in [GUIDELINES.md](../../GUIDELINES.md); OSS practices adopted from sibling projects per [oss-playbook.md](../planning/oss-playbook.md). All new/amended ADRs remain **Proposed pending operator review**. |
