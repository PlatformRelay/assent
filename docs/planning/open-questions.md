# Open questions

| ID | Question | Blocks | Notes / leading answer |
| --- | --- | --- | --- |
| OQ-1 | **Project name** (and module path, org). Candidates: [naming.md](naming.md) | any public repo (D-001) | operator shortlist pending |
| OQ-2 | Hosting: GitHub only, or GitLab mirror (dogfooding the GitLab adapter on our own repo)? | Phase 5 / E9 | dogfooding on GitLab is attractive once E4 exists |
| OQ-3 | ~~Two parallel frontends?~~ Resolved by ADR-0002 v2: one YAML envelope, pluggable predicate backends | — | superseded; successor questions: OQ-11/OQ-12 |
| OQ-4 | Ship gRPC (`go-plugin`) tier in v1, or is HTTP/exec enough alongside built-ins? | ADR-0004 accept | Spike C; leading: defer gRPC to v1.x |
| OQ-5 | Policy discovery: `.verdict/` only, or also remote packs (central policy repo, git-ref-pinned) in v1? | ADR-0010 accept | leading: local v1, remote packs designed-for (ADR-0010) |
| OQ-6 | E2E default in CI: GitLab-in-kind vs GitLab CE testcontainer (boot time / RAM / flakiness)? | ADR-0006 accept | Spike B; kind stays for local/demo either way |
| OQ-7 | GitHub mapping for `challenge`: `REQUEST_CHANGES` + required-conversation-resolution — sufficient parity with GitLab's all-discussions-resolved gate? | ADR-0005 accept | write the dossier in Phase 1.3 |
| OQ-8 | Decision replay/audit: JSON report artifact enough, or signed/attested decision record later? | Phase 3 | v1: artifact (Pins in report); attestations later epic |
| OQ-9 | Version pinning for reproducibility (tool digest + policy SHA in report `Pins`)? | Phase 3 | must be in the report schema from day 1 |
| OQ-10 | Monorepo support: multiple policy scopes per repo (path-scoped `.verdict/` dirs)? | Phase 3 | likely bindings-level path scoping |
| OQ-11 | `assert` backend implementation: embed **kyverno-json** (`pkg/jsonengine`) vs native **cel-go** — maturity vs control? | ADR-0002 accept | Spike A; wrapper interface makes it reversible either way |
| OQ-12 | `assert` authored syntax: Kyverno assertion trees, CEL expression strings, or trees-with-CEL-leaves? | ADR-0002/0010 accept | Spike A decides by writing all archetypes in each |
| OQ-13 | Risk score conventions: point scale, per-binding thresholds only, or also effect escalation (env promotes `challenge`→`block`)? | ADR-0007 accept | start: points + thresholds only |
| OQ-14 | `serve` (webhook) in v1 or v1.x? Event dedup + re-eval-on-thread-resolution semantics needed | ADR-0009 accept | leading: v1.x, architecture-ready from day 1 |
| OQ-15 | Resource-rename detection: similarity threshold, and is fold-to-`rename` opt-in per class? | ADR-0003 accept | fail-safe fallback = raw delete+add (stricter) |
| OQ-16 | Which **open-source repos** join the demo/test corpus (kubernetes/org, JulieOps/kafka-gitops topologies, octoDNS zones, Backstage catalogs, GitHub safe-settings)? | Phase 1.1 | operator also provides 2–3 generalized private shapes (D-008) |
