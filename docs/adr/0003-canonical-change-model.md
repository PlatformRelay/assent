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

### Deletions and renames are first-class

- **Entry deletion** (a map key / resource removed within a file) emits `delete` with the full
  old value — so "topic is being deleted" is a plain match, and packs can attach `block` or
  `challenge` effects to it (never silently folded into a modify).
- **File events** (`added | deleted | renamed | modified | opaque`) are tracked alongside
  field changes; file renames use git rename detection and preserve identity (`from`/`to`), so
  a rename is not reported as delete-everything + add-everything.
- **Resource renames** (entry key changes but the value is identical/similar) are detected
  heuristically within a file: a `delete`+`add` pair with equal (or near-equal) values is
  folded into `rename {oldPath, newPath}`. Policies decide what a rename means (often:
  `challenge` — renames of live resources are usually destructive downstream). The heuristic's
  similarity threshold and its failure mode (fall back to the raw delete+add pair, which is
  *stricter*) must be golden-tested exhaustively — spec'd in Phase 3.

## Consequences

- Policies are portable across formats and repos; the ChangeSet schema joins PolicyInput as a
  frozen public contract.
- Adding a format (TOML, properties, …) = one adapter + conformance fixtures, no policy changes.

## Counterpoints considered

- *"JSON-merge-patch or JSONPatch already exist."* — JSONPatch is a good serialization
  candidate for `Change`, but alone it lacks old-values, file events, and opaque fallbacks;
  we may still adopt its path syntax (RFC 6901).

## Amendment (2026-07-21, adversarial review F12)

Fold-to-`rename` is **opt-in per class** (`classes[].renames: detect|raw`, default `raw`),
because the similarity threshold is otherwise an attacker-tunable downgrade knob (craft the
paired add to sit just above the threshold and convert a `block`-able delete into a
resolvable `challenge`). Additionally, a `rename` can never be treated *less* strictly than
the `delete` of the same class: the engine applies the stricter of the class's delete/rename
effects. Golden tests must include adversarial near-threshold pairs, not only correctness
pairs.

## Amendment 2 (2026-07-21, security review A-05 / review P2-11)

- **Input limits**: format adapters enforce max file size/count, nesting depth, and YAML
  anchor/alias expansion caps (billion-laughs); symlinks and path-traversal names are
  rejected; parse runs under a deadline. Any breach yields `opaque-change` → fail-safe
  REVIEW (never a crash, never a skip).
- **Source positions are first-class**: every `Change` carries file + line/column spans for
  old and new values (adapters must preserve positions at parse time — retrofitting them
  later means rewriting the parsers). This is what makes forge inline/line-anchored comments
  possible; `Finding.Paths` alone cannot anchor a thread to a diff line.
