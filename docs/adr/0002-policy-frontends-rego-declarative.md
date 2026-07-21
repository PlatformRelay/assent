# ADR-0002: Policy frontends: Rego + Kyverno-style declarative YAML over one engine

| | |
| --- | --- |
| **Status** | Proposed |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0003 change model](0003-canonical-change-model.md) · [ADR-0004 plugins](0004-plugin-architecture.md) |

## Context

Adopters must express merge policies themselves. Two audiences exist: policy engineers who
already speak **Rego** (OPA is the lingua franca of policy-as-code) and platform teams who
prefer **declarative YAML** in the style of Kyverno — pattern-match on the change, state the
constraint, no programming. Forcing either audience into the other's syntax loses half the
market. Both must be deterministic, testable, and evaluate against the same input document.

## Options

| Option | Pros | Cons |
| --- | --- | --- |
| Rego only | maximal power, OPA tooling (`opa test`, coverage) for free | steep for casual adopters; "wall of Rego" scares small repos |
| Declarative YAML only | lowest entry barrier | hits expressiveness ceiling fast (cross-entry checks, arithmetic); we end up inventing a bad programming language in YAML |
| **Both, layered: YAML frontend compiles/lowers to the same evaluation semantics as Rego; one engine, one input contract** | each audience served; YAML rules stay simple *because* the escape hatch to Rego exists; single decision/finding model | two surfaces to document and test; YAML-to-engine lowering must be spec'd precisely |
| CEL expressions embedded in YAML | familiar from k8s ValidatingAdmissionPolicy | still a third syntax; doesn't remove the need for either full Rego or full YAML patterns |

## Decision (proposed)

**Both frontends over one engine.** The engine consumes a single canonical **PolicyInput**
document (change model + facts + MR metadata, [ADR-0003](0003-canonical-change-model.md)) and
produces the same **Decision + Findings** contract regardless of frontend. The declarative
YAML frontend covers the ~80% archetypes (ownership, bounded change, allow-listed fields,
no-destruction, environment split); Rego is the escape hatch for the rest. A YAML rule and a
Rego rule are indistinguishable downstream.

Open sub-questions for the planning phase: exact YAML schema (Kyverno-inspired
`match`/`validate` vs. custom), whether YAML lowers *to generated Rego* (one evaluator) or to
a native evaluator (two evaluators, one contract), and how `opa test` integrates with the
built-in fixture harness.

## Consequences

- The PolicyInput document schema becomes the most important public contract of the project —
  spec'd first, versioned, and frozen per major version.
- Every example in `examples/` must exist in **both** syntaxes where expressible, as living
  documentation of the equivalence.

## Counterpoints considered

- *"Two frontends = double maintenance."* — Mitigated if YAML lowers to Rego (single
  evaluator); this is the leading sub-option and will be decided by spike + trade-off matrix.
