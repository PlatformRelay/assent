# Local kind cluster for e2e / demo

Hosts the test GitLab instance ([test/e2e](../../test/e2e/README.md), path 1).

```bash
kind create cluster --name assent --config kind-config.yaml
# GitLab install + seeding scripts arrive with meta-plan Spike B / epic E7.
```

Planned contents: GitLab install (Helm values or CE-in-pod manifest), a `seed/` script that
creates groups, users, bot token, sample projects from `examples/repos/`, and fixture MRs.
