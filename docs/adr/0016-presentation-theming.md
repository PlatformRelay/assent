# ADR-0016: Presentation theming — config knobs, slots, CEL messages, render contract

| | |
| --- | --- |
| **Status** | Accepted (P2-E5 / D-015) |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0012 presentation](0012-presentation-templates-debug.md) (supersedes its override mechanism) · [ADR-0013 assert/CEL](0013-assert-syntax-and-backend.md) · [ADR-0014 test format](0014-adopter-test-format.md) · [ADR-0015 §1](0015-trust-boundaries-merge-integrity.md) · D-012 · D-015 · design roast 2026-07-21 P1-6/P2-2 |

## Context

ADR-0012 fixed the right spine (structured findings → central renderer → forge markdown;
escaping/redaction in the renderer) but exposed flexibility as whole-file Go `text/template`
overrides. Critique that prompted this ADR: (a) whole-file overrides fork the default and can
destroy the lifecycle markers idempotence depends on; (b) rule authors would face three
expression dialects (CEL in `assert`, `{{ }}` interpolation, Go templates in overrides), and
Go templates fail silently (`<no value>`) — in the tool whose product is explanation;
(c) the debug story deserves a rendered artifact, not just JSON; (d) customizable
presentation was untestable; (e) the template-visible data shape was an unpinned contract.

## Decision (proposed)

### 1. The renderer owns the envelope; customization fills slots

Lifecycle markers (decision hash, finding keys), escaping, redaction, and length-clamping
are applied by the renderer **outside** any user-customizable region — invariant, enforced
by construction (templates never see or emit markers). Customization is tiered, mirroring
the policy-surface philosophy:

- **Tier 0 — config knobs** (`.assent/config.yaml → presentation:`): `verbosity:
  minimal|standard|full` (global and per-environment — dev chatty, prod terse), emoji
  on/off, collapse threshold for high-cardinality findings, `locale`.
- **Tier 1 — slot overrides** (`.assent/templates/`): named regions per artifact —
  `headline`, `docs`, `details`, `footer` — individually replaceable; untouched slots keep
  receiving default-theme improvements.
- **Tier 2 — full artifact template**: escape hatch per artifact type (finding thread,
  summary comment, explain output, markdown report).

Per D-012 adoption gating: **v1 ships tier 0 + CEL messages + the envelope invariant +
render/golden testing**; tiers 1–2 and the markdown report artifact unlock with the first
named consumer who needs them (the seams are designed; no frozen template contract until
then). Templates load from the target ref only (ADR-0015 §1).

### 2. One expression language: `{{ }}` wraps CEL

Message interpolation in `message`, `docs.summary`, `debug:` lines, and slot templates is
**CEL over the rule's predicate scope** (`old`, `new`, `entry`, `facts`, `env`, `mr` — the
ADR-0013 appendix table), following the K8s ValidatingAdmissionPolicy `messageExpression`
precedent. One activation model serves `assert` and messages (already Spike A scope).
Unknown fields or type errors are **load-time lint errors** — never `<no value>` at
render time. Go `text/template` may remain the internal engine of the default theme; it is
no longer an authored surface.

### 3. PresentationModel is the pinned render contract (roast P1-6)

The record splits four ways — rendered markdown never participates in decision identity:

| Record | Content | Audience |
| --- | --- | --- |
| `DecisionRecord` | redacted, stable outcome + evidence digests | audit, `stats`, report artifact |
| `ReplayBundle` | access-controlled canonical input (incl. sensitive facts, protected) | hermetic replay |
| `PresentationModel` | redacted, renderer-only view of findings/trace | templates, `explain`, report.md |
| `PublicationReceipt` | forge state + operations performed | reconciliation, debugging |

The `PresentationModel` schema is a **versioned public contract** (it is what slot/full
templates and `explain` consume); it freezes in Phase 3 alongside the others. Exact shapes
of all four records are settled with the 2026-07-21 design-roast processing.

### 4. Rendering is testable

- `assent render --finding <fixture> [--template-dir …]` previews any artifact from a
  fixture without a live MR.
- The default theme carries **golden markdown snapshot tests** in this repo — "strong
  defaults" is enforced, not aspirational.
- Template lint: unknown slot, unknown field, marker-region violation, unescaped-raw usage —
  load-time errors.
- Separation of concerns (roast P2-2): adopter **policy tests** (ADR-0014) assert structured
  safety semantics (rule, effect, paths — not wording); **template/theme tests** assert
  rendered markdown via `render` goldens. A wording change never breaks a safety test.

### 5. Chrome strings are a locale catalog

Fixed renderer strings ("Resolve this thread to confirm", "Evaluation details", expiry
notices) live in a string catalog keyed by `presentation.locale`; v1 ships `en` only, but
translation is a data contribution, not a template fork. Rule-authored messages are the
pack author's language — untouched.

## Consequences

- ADR-0012's "repo-level template overrides in `.assent/templates/` — Go text/template" is
  superseded by §1–2; everything else in ADR-0012 (default layout, lifecycle state machine,
  redaction) stands.
- ADR-0014 gains the render-vs-safety test split (amendment); `message~:` assertions in
  policy tests are discouraged in favor of structured finding assertions.
- Spike A adds: CEL-as-interpolation in one activation model; render golden harness.
- Cost honestly stated: slots + PresentationModel add surface. Accepted because presentation
  is v1 critical path, tier 0 covers most users, and tiers 1–2 are adoption-gated (D-012).

## Counterpoints considered

- *"Just document 'copy the default template and edit it'."* — That is exactly the fork
  that rots: dropped markers, missed improvements, unescaped injections. The envelope
  invariant is only enforceable if user content cannot reach it.
- *"CEL in messages is code in strings again."* — Yes, the same code the author already
  wrote in `assert`, checked at the same load time, with the same scope. One dialect
  beats two half-known ones.
