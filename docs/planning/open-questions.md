# Open questions

| ID | Question | Blocks | Notes / leading answer |
| --- | --- | --- | --- |
| OQ-1 | ~~Project name~~ **Resolved: assent** (D-009); repo live at PlatformRelay/assent (D-014). Residual: `assent.dev` domain unverified — decide apiVersion group before Phase 3 freeze | Phase 3 | [naming.md](naming.md) |
| OQ-2 | Hosting: GitHub only, or GitLab mirror (dogfooding the GitLab adapter on our own repo)? | Phase 5 / E9 | dogfooding on GitLab is attractive once E4 exists |
| OQ-3 | ~~Two parallel frontends?~~ Resolved by ADR-0002 v2: one YAML envelope, pluggable predicate backends | — | superseded; successor questions: OQ-11/OQ-12 |
| OQ-4 | Ship gRPC (`go-plugin`) tier in v1, or is HTTP/exec enough alongside built-ins? | ADR-0004 accept | Spike C; leading: defer gRPC to v1.x |
| OQ-5 | Policy discovery: `.assent/` only, or also remote packs (central policy repo, git-ref-pinned) in v1? | ADR-0010 accept | leading: local v1, remote packs designed-for (ADR-0010) |
| OQ-6 | E2E default in CI: GitLab-in-kind vs GitLab CE testcontainer (boot time / RAM / flakiness)? | ADR-0006 accept | Spike B; kind stays for local/demo either way |
| OQ-7 | GitHub mapping for `challenge`: `REQUEST_CHANGES` + required-conversation-resolution — sufficient parity with GitLab's all-discussions-resolved gate? | ADR-0005 accept | write the dossier in Phase 1.3 |
| OQ-8 | Decision replay/audit: JSON report artifact enough, or signed/attested decision record later? | Phase 3 | v1: artifact (Pins in report); attestations later epic |
| OQ-9 | Version pinning for reproducibility (tool digest + policy SHA in report `Pins`)? | Phase 3 | must be in the report schema from day 1 |
| OQ-10 | Monorepo support: multiple policy scopes per repo (path-scoped `.assent/` dirs)? | Phase 3 | likely bindings-level path scoping |
| OQ-11 | ~~kyverno-json vs cel-go~~ **Leading answer: cel-go** (ADR-0013 — kyverno-json dormant ~18mo, bus factor ≈1, heavy deps) | ADR-0013 accept | Spike A narrowed to residual code risk (numeric coercion, error UX, cost/purity, trace wiring, activation model) |
| OQ-12 | ~~assert authored syntax~~ **Leading answer: hybrid** — `all`/`any`/`not` trees with CEL leaves + per-leaf `message`, string shorthand (ADR-0013 + gallery) | ADR-0013 accept | existing draft samples stay valid as shorthand |
| OQ-13 | Risk score conventions: point scale, per-binding thresholds only, or also effect escalation (env promotes `challenge`→`block`)? | ADR-0007 accept | start: points + thresholds only |
| OQ-14 | `serve` (webhook) in v1 or v1.x? Event dedup + re-eval-on-thread-resolution semantics needed | ADR-0009 accept | leading: v1.x, architecture-ready from day 1 |
| OQ-15 | ~~fold-to-rename opt-in?~~ **Committed: opt-in per class, default raw; rename never laxer than delete** (ADR-0003 amendment, review F12) | ADR-0003 accept | residual: similarity metric itself |
| OQ-17 | `facts.max_age` default (ADR-0015 §3): 24h reasonable? Per-provider overrides? | ADR-0015 accept | security-sensitive facts likely need shorter |
| OQ-18 | GitHub challenge parity (sharpens OQ-7 after ADR-0009 amendment): can REQUEST_CHANGES + conversation-resolution + SHA-pinned auto-merge reproduce the GitLab arm-and-wait flow, incl. who dismisses the bot's review? | ADR-0005 accept / E8 | Phase 1.3 dossier |
| OQ-19 | Post-merge reconciliation: bad-automerge audit + auto-revert-MR creation — v1.x epic or out of scope? (Second review flagged: `scan` replays but detects nothing post-merge) | roadmap | needs serve/scheduled mode; design seam only for now |
| OQ-20 | Batch/sweep apply mode (evaluate+act over many open MRs in one run, cron-style) — or is per-MR CI + serve enough? | roadmap | `scan` stays recorder-only either way |
| OQ-21 | Per-rule rollout phases (off / advisory / enforce) in the envelope — or is editing effects in the pack (advisory = comment) sufficient? | Phase 3 | leading: no phase field in v1; document the effect-editing pattern |
| OQ-22 | Envelope `match` on MR metadata: labels, draft status, author allowlists — which belong in `match.mr` for v1? | Phase 3 | draft-MRs likely skipped by default in CI template |
| OQ-16 | Which **open-source repos** join the demo/test corpus (kubernetes/org, JulieOps/kafka-gitops topologies, octoDNS zones, Backstage catalogs, GitHub safe-settings)? | Phase 1.1 | operator also provides 2–3 generalized private shapes (D-008) |
