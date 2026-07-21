# Decision log

Lightweight, dated operator decisions (D-nnn). ADRs cover architecture; this file covers
project/process decisions.

| ID | Date | Decision |
| --- | --- | --- |
| D-001 | 2026-07-21 | Working title **verdict-2**; final name undecided (candidates: [naming.md](../planning/naming.md)). **No public GitHub repo until the name is chosen.** Local git only for now. |
| D-002 | 2026-07-21 | License **Apache-2.0** (Kubernetes/Argo CD precedent: permissive + patent grant; supersedes the initial MIT choice same day). Fully open source; no employer names, internal system names, or internal policy content in any committed artifact. |
| D-003 | 2026-07-21 | Go module path `github.com/PlatformRelay/verdict2` is a **placeholder**; revisit together with D-001 before first push. |
| D-004 | 2026-07-21 | Git identity for this repo: `Konrad Heimel <konrad.heimel@gmail.com>` (local git config). No AI co-author trailers. |
| D-005 | 2026-07-21 | Development is **spec/test-driven**: openspec workflow, REQ IDs with `Test:` + `Verify:`, TDD for all implementation work. |
| D-006 | 2026-07-21 | Operator preference: **Kyverno-style declarative YAML is the primary policy surface**; Rego is an escape-hatch predicate backend inside the same envelope, never a parallel frontend (ADR-0002 v2). |
| D-007 | 2026-07-21 | **No database for now**: the per-run JSON report artifact is the storage format; `stats` aggregates report files (ADR-0009/0012). |
| D-008 | 2026-07-21 | Sample corpus: operator provides 2–3 further real self-service repo shapes (generalized, never verbatim); additionally curate **open-source repos** as public test/demo corpora (candidates in [examples/repos](../../examples/repos/README.md)). |
