# Sample self-service repos

Generic, **generated** sample repos (never copied from any private codebase — D-002). They
serve three purposes: e2e seeds (created in the test GitLab, see `test/e2e/`), adopter
documentation, and fixtures for the policy test harness.

Planned shapes (built in meta-plan Phase 1.1, refined as the operator provides real-world
repo shapes to generalize):

| Sample | Format | Exercises |
| --- | --- | --- |
| `topic-registry/` | YAML (one file per topic: name, owner, partitions, retention) | ownership, bounded change, no-destruction |
| `service-catalog/` | JSON (single catalog file, many entries) | allow-listed fields, schema validity, multi-entry diffs |
| `infra-vars/` | tfvars/HCL (per-env variable files) | environment split, HCL parsing, opaque-change fallback |
