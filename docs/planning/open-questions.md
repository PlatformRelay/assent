# Open questions

| ID | Question | Blocks | Notes / leading answer |
| --- | --- | --- | --- |
| OQ-1 | **Project name** (and module path, org). | any public repo (D-001) | brainstorm session pending |
| OQ-2 | Hosting: GitHub only, or GitLab mirror (dogfooding the GitLab adapter on our own repo)? | Phase 5 / E9 | dogfooding on GitLab is attractive once E4 exists |
| OQ-3 | Declarative YAML frontend: lower to generated Rego (one evaluator) or native evaluator (two evaluators, one contract)? | ADR-0002 accept | Spike A; leading: lower to Rego |
| OQ-4 | Ship gRPC (`go-plugin`) tier in v1, or is HTTP/exec enough alongside built-ins? | ADR-0004 accept | Spike C; leading: defer gRPC to v1.x |
| OQ-5 | Policy discovery in governed repos: `.verdict/` dir? central policy repo reference? both (remote packs + local overrides)? | Phase 3 contracts | multi-repo orgs will want central packs with per-repo pinning |
| OQ-6 | E2E default in CI: GitLab-in-kind vs GitLab CE testcontainer (boot time / RAM / flakiness)? | ADR-0006 accept | Spike B; kind stays for local/demo either way |
| OQ-7 | GitHub mapping for "must-resolve findings": `REQUEST_CHANGES` review + conversation resolution — is required-conversation-resolution branch protection sufficient parity with GitLab's all-discussions-resolved gate? | ADR-0005 accept | write the dossier in Phase 1.3 |
| OQ-8 | Decision replay/audit: is the JSON report artifact per run enough, or do we need a signed/attested decision record (SLSA-style) for compliance-minded adopters? | Phase 3 | v1: artifact; attestations as later epic |
| OQ-9 | How do adopters pin policy + tool versions for reproducibility (tool image digest + policy git SHA in the report)? | Phase 3 | must be in the report schema from day 1 |
| OQ-10 | Monorepo support: multiple policy scopes per repo (path-scoped policy dirs)? | Phase 3 | likely `match.paths` at pack level |
