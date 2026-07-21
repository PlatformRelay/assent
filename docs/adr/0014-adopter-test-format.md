# ADR-0014: Adopter test format — policy tests as a public contract

| | |
| --- | --- |
| **Status** | Proposed |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0006 testing](0006-testing-strategy.md) · [ADR-0007 effects](0007-rule-effects-decision-aggregation.md) · [ADR-0010 config](0010-config-files-repo-layout.md) · D-010 |

## Context

Adopters will not trust automerge they cannot test. The fixture format under
`.assent/tests/` is therefore not tooling detail — it is a **frozen public contract**
(meta-plan Phase 3) on par with PolicyInput, and the primary vehicle for descriptive,
example-driven policy development in any self-service repo. Design goals: readable as
documentation, writable without knowing Go, diffable in review, and expressive enough to
pin decisions, findings, and score arithmetic.

## Decision (proposed)

### Directory case (full form)

```
.assent/tests/<pack>/<case-name>/
├── given/
│   ├── base/…        # files as on the target branch (may be empty for "new file" cases)
│   ├── head/…        # the same files as the MR proposes them
│   ├── facts.yaml    # stubbed provider results (providers are never called in tests)
│   └── mr.yaml       # optional MR metadata: author, labels, target branch (defaults exist)
└── expect.yaml
```

The engine derives the ChangeSet from `base/` vs `head/` with the *production* differ and
classifier — tests exercise the real pipeline (integration-level by construction), with only
providers and the forge stubbed.

### `expect.yaml`

```yaml
decision: REVIEW                 # APPROVE | REVIEW | BLOCK  (required)
findings:                        # must-contain by default; `exact: true` for closed lists
  - rule: topics/retention-shrink-challenge
    effect: challenge
    path: "/retentionMs"
    message~: "data loss"        # `~` suffix = substring/regex match on rendered message
absent:                          # rules that must NOT fire
  - topics/no-topic-deletion
score: { total: 3, threshold: 4 }   # optional, pins the arithmetic
```

### Inline shorthand (single-file cases)

For the common "one field changed" case, a single YAML file
(`.assent/tests/<pack>/cases.yaml`) holds many small cases:

```yaml
cases:
  - name: partition-increase-ok
    file: topics/prod/orders.yaml
    base: { name: orders, owner: team-a, partitions: 12 }
    head: { name: orders, owner: team-a, partitions: 24 }
    facts: { quota: { max_partitions: 24 }, author: { groups: [team-a] } }
    expect: { decision: APPROVE }
```

### Runner semantics (`assent test`)

- Runs every case; failure output shows expected vs actual decision, the finding diff, and a
  ready-to-copy actual block (ADR-0012 hints style, cf. walkthrough).
- `--update` golden-flow: writes actuals into `expect.yaml` for review-by-diff.
- **Rule coverage**: every rule must be exercised by ≥1 case *where its predicate holds* and
  (for `vouch` rules) ≥1 where it does not; `assent lint` fails otherwise. Coverage report
  via `assent test --coverage`.
- Determinism: each case runs twice, results must be identical (same gate as golden L0 tests).
- CI templates run `assent test` on every MR that touches `.assent/` — policies gate
  themselves.

## Consequences

- Fixtures double as documentation and as our own e2e seeds (`examples/` packs must keep
  their fixtures green — dogfooding, ADR-0006).
- The format needs a JSON schema + versioned `apiVersion` like every other contract; changes
  go through openspec proposals.
- `--update` makes golden maintenance cheap but demands review discipline — the CI template
  therefore always runs tests from the *target* branch's expectations when policies change
  (interaction with the policy-ref question raised by the security review — resolve together).

## Counterpoints considered

- *"Reuse `opa test` / Rego unit tests."* — Covers only Rego-backend rules, tests predicates
  in isolation rather than the envelope+aggregation pipeline, and is unwritable for the
  YAML-first audience. The harness may still *run* `opa test` additionally for rego/ modules.
- *"Table-driven Go tests."* — Our internal tests, yes (D-010); adopters never write Go.
