# ADR-0006: Testing strategy: spec-driven pyramid with real-forge e2e

| | |
| --- | --- |
| **Status** | Proposed |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0005 forge](0005-forge-abstraction-gitlab-first.md) · [openspec/config.yaml](../../openspec/config.yaml) |

## Context

Two distinct testing obligations exist: (a) **our own** correctness — engine, frontends,
adapters — and (b) the **adopter's** ability to test *their* policies descriptively. A merge
gate that adopters cannot test will not be trusted with automerge. E2E must exercise a real
forge, because thread/approval/merge semantics are exactly what mocks get wrong.

## Decision (proposed)

**Spec-first pyramid; every spec REQ carries `Test:` (artifact path) and `Verify:` (command).**

| Level | Scope | Vehicle |
| --- | --- | --- |
| L0 | engine, change model, frontends | Go unit + **golden decision tests** (fixture PolicyInput → expected Decision/Findings, diffable JSON) + property tests for the differ |
| L1 | policy packs | the **user-facing harness**: `verdict2 test` runs fixture changes against a policy dir and asserts decisions — same harness we ship to adopters, dogfooded on `examples/` |
| L2 | forge adapters | contract tests against recorded API cassettes |
| L3 | end-to-end | real GitLab: **kind-hosted GitLab** (local, `hack/kind/`) and/or **GitLab testcontainer** (CI) seeding generated sample repos; conformance suite per ADR-0005; GitHub via test org later |

Open trade-off (decide by spike, tracked as OQ-6): GitLab-in-kind (heavier, closer to a
long-lived shared instance, doubles as demo environment) vs. GitLab CE testcontainer
(self-contained per CI run, slow image boot ~minutes, memory-hungry). Both paths are scaffolded
under `test/e2e/`; the spike measures boot time and flakiness and picks the CI default.

**Determinism gate**: golden tests re-run each decision twice and diff — any nondeterminism is
a build failure.

## Consequences

- The adopter harness is a first-class product surface, not an afterthought — it gets its own
  spec epic and its own examples.
- Sample repos are *generated* (scripts under `test/e2e/`), never copied from any private
  codebase; they double as documentation.

## Counterpoints considered

- *"Mock the forge for e2e, it's cheaper."* — Rejected: resolvable-thread and approval
  semantics are the product's riskiest integration surface; mocks would test our assumptions,
  not the forge.
