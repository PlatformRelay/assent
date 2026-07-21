# ADR-0010: Configuration files and governed-repo layout

| | |
| --- | --- |
| **Status** | Accepted (partial: prove/require envelope shapes per ADR-0017 §2/§5; P2-E5) |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0002 policy surface](0002-policy-frontends-rego-declarative.md) · [ADR-0007 effects](0007-rule-effects-decision-aggregation.md) · [ADR-0008 routing](0008-change-classification-routing-scope.md) · OQ-5 |

## Context

Adopters interact with the tool almost exclusively through files in their repo. The file
taxonomy *is* the UX. It must separate concerns cleanly: repo wiring (environments, classes,
providers) vs. routing (bindings, thresholds) vs. rules (packs) vs. tests — so that packs are
shareable across repos while wiring stays local.

## Decision (proposed)

Everything lives under **`.assent/`** in the governed repo (name follows OQ-1):

```
.assent/
├── config.yaml          # repo wiring: environments, change classes, providers
├── bindings.yaml        # routing: (class, environment) -> packs + risk thresholds
├── packs/
│   └── topics/
│       ├── pack.yaml    # pack metadata (name, version, description)
│       ├── rules/       # MergePolicy documents (YAML envelope, ADR-0002)
│       └── rego/        # escape-hatch Rego modules referenced by rules
└── tests/
    └── topics/
        └── partition-increase-ok/
            ├── given/   # fixture: changed files (base/ and head/ variants) + facts.yaml
            └── expect.yaml  # expected decision + findings
```

Remote packs (central policy repos, pinned by git ref) are planned via
`packs: [git::https://…//packs/topics?ref=v1.2.0]` in bindings — local overrides win (OQ-5).

### `config.yaml` — repo wiring

```yaml
apiVersion: assent.dev/v1alpha1
kind: Config
environments:
  - name: prod
    match: { paths: ["topics/prod/**", "envs/prod/**"] }
  - name: dev
    match: { paths: ["**"] }   # last match wins as default
classes:
  - name: kafka-topic
    match: { paths: ["topics/**/*.yaml"] }
  - name: infra-vars
    match: { paths: ["**/*.tfvars"] }
providers:
  author:                        # -> facts.author.*
    type: builtin/gitlab-groups
  quota:                         # -> facts.quota.*
    type: http
    url: https://quota.example.com/api/v1/lookup
    failure: closed              # closed (default) -> REVIEW; open -> skip facts
```

### `bindings.yaml` — routing + risk

```yaml
apiVersion: assent.dev/v1alpha1
kind: RulesetBinding
bindings:
  - class: kafka-topic
    environment: dev
    packs: [topics]
    risk: { threshold: 10 }
  - class: kafka-topic
    environment: prod
    packs: [topics, topics-strict]
    risk: { threshold: 4 }
  - class: infra-vars
    environment: "*"
    packs: [tfvars]
    risk: { threshold: 6 }
```

### A rule file — envelope with effects, scope, both predicate backends

```yaml
apiVersion: assent.dev/v1alpha1
kind: MergePolicy
metadata: { name: topic-safety }
spec:
  rules:
    - name: partition-increase-within-quota   # tier 1: assert predicate
      match: { changes: [{ path: "**/partitions", kind: modify }] }
      assert: "new >= old && new <= facts.quota.max_partitions"
      effect: vouch
      points: 1
    - name: retention-shrink-challenge
      match: { changes: [{ path: "**/retentionMs", kind: modify }] }
      assert: "new < old"        # predicate true -> effect applies
      effect: challenge
      message: "Retention shrinks from {{ old }} to {{ new }} — data loss possible. Sure?"
    - name: naming-convention                 # tier 2: rego, branch scope (ADR-0008)
      match: { classes: [kafka-topic] }
      scope: branch
      rego: { file: ../rego/naming.rego }     # returns findings data only
      effect: comment
      points: 2
    - name: no-topic-deletion
      match: { changes: [{ path: "topics/**", kind: delete }] }
      effect: block
      message: "Topic deletion is never auto-mergeable."
```

## Consequences

- `config.yaml` is the only file that knows company-specific wiring (providers!); packs stay
  portable and publishable. This is the seam that makes per-company permission
  reimplementation (ADR-0004) a config exercise plus one small provider service.
- Tests are first-class repo citizens next to the packs they test; `assent lint` fails packs
  without tests.
- All kinds share one `apiVersion` line for engine-version gating and future migrations.

## Counterpoints considered

- *"One big file is simpler."* — For toy repos, yes; it destroys pack shareability and makes
  ownership (CODEOWNERS on `.assent/packs/x/`) impossible. `init` can still generate a
  minimal single-pack layout.

## Amendment (2026-07-21, second review P2-9 / security review A-12)

- Starter packs must match destructive intent, not one change kind: `no-topic-deletion`-style
  rules match `kind: [delete, rename]` — otherwise the rename-fold (ADR-0003) walks a
  de-facto delete+recreate past the block rule.
- Remote packs (when they land, OQ-5): pinned by **commit SHA** (tags are mutable), checksum/
  signature verified, and subject to the same target-ref/no-self-modification rule as local
  policy (ADR-0015 §1).
- The lint hard-error list (vouch scoping, reserved classes, environment priority,
  fail-open restrictions, docs-on-challenge/block, tests-per-rule) is consolidated in the
  Phase 3 spec for `assent lint`.
