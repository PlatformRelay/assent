# C4 — Level 1: System context

```mermaid
C4Context
    title assent — system context

    Person(contributor, "Contributor", "Opens MRs/PRs against a self-service config repo")
    Person(platform, "Platform engineer", "Owns the repo; authors and tests policies")

    System(assent, "assent", "Deterministic policy-driven auto-merge gate, executed as a CI job per MR/PR")

    System_Ext(forge, "Forge", "GitLab (first) / GitHub (next): hosts repo, MR/PR, threads, approvals, merge")
    System_Ext(ci, "CI runner", "GitLab CI / GitHub Actions: triggers assent per MR event")
    System_Ext(idp, "Permission sources", "Keycloak / LDAP / forge groups / ownership files — via provider plugins")
    System_Ext(facts, "Fact sources", "Site-specific systems answering context questions — via provider plugins")

    Rel(contributor, forge, "opens MR / pushes changes")
    Rel(platform, forge, "maintains repo + policy dir (.assent/)")
    Rel(ci, assent, "runs per MR/PR")
    Rel(assent, forge, "reads diff & metadata; posts threads/comments; approves/denies; merges")
    Rel(assent, idp, "resolves author permissions")
    Rel(assent, facts, "resolves external facts")
```

Key property: assent is **stateless per invocation** — every run recomputes the decision
from (diff, repo snapshot, facts, policy version). No database, no long-lived service in v1.
