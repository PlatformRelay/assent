# ADR-0003: Canonical change model for JSON / YAML / HCL-tfvars

| | |
| --- | --- |
| **Status** | Proposed |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0002 policy frontends](0002-policy-frontends-rego-declarative.md) |

## Context

Policies must reason about *what changed semantically*, not about diff hunks. The tool targets
repos holding JSON, YAML, and HCL/tfvars. A YAML key reordering or comment edit is a no-op; a
nested field flipping from `3` to `1` is a bounded-change question; a removed map entry is a
deletion. Raw `git diff` cannot express any of this.

## Options

| Option | Pros | Cons |
| --- | --- | --- |
| **Parse base & head per file into a generic value tree; structural diff → field-level ChangeSet (add/modify/delete with JSON-Pointer-style paths, old/new values)** | format-agnostic policies; trivially serializable as PolicyInput; no-op edits (comments, ordering) disappear | needs per-format parser adapters; HCL expressions (non-literal) need a defined representation |
| Line-diff + per-rule parsing | no upfront model | every rule reinvents parsing; non-deterministic corner cases |
| Format-specific models (one per file type) | precise per format | policies stop being portable across repos/formats |

## Decision (proposed)

**Generic value-tree diff → canonical ChangeSet.** Format adapters (JSON, YAML via mapping to
the same node type, HCL/tfvars via `hashicorp/hcl`) parse into one value tree; a structural
differ emits `Change{path, kind: add|modify|delete, old, new}` entries plus file-level events
(file added/deleted/renamed). Unparseable or unknown files surface as explicit
`opaque-change` events that policies must handle (default: require human review — fail safe).

HCL caveat to spec precisely: tfvars are literal-only (easy); full HCL with expressions is
represented but expression *evaluation* is out of scope for v1.

## Consequences

- Policies are portable across formats and repos; the ChangeSet schema joins PolicyInput as a
  frozen public contract.
- Adding a format (TOML, properties, …) = one adapter + conformance fixtures, no policy changes.

## Counterpoints considered

- *"JSON-merge-patch or JSONPatch already exist."* — JSONPatch is a good serialization
  candidate for `Change`, but alone it lacks old-values, file events, and opaque fallbacks;
  we may still adopt its path syntax (RFC 6901).
