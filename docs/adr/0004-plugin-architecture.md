# ADR-0004: Plugin architecture for permission & fact providers

| | |
| --- | --- |
| **Status** | Proposed |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0001 language](0001-language-go-single-binary.md) · [ADR-0002 policy frontends](0002-policy-frontends-rego-declarative.md) |

## Context

Rules frequently need **external answers**: "is the author in the owning group?" (Keycloak,
LDAP, GitLab/GitHub groups), "does this cost center exist?", "is this service registered?".
These providers are inherently site-specific — the whole point of the project is that a company
can wire its own without forking. Unlike a first-party tool, we **must** assume third-party,
possibly untrusted, possibly non-Go plugin authors. Policies themselves are already
runtime-loaded data (ADR-0002); this ADR is about *imperative* extension points.

## Options

| Option | Pros | Cons |
| --- | --- | --- |
| Compile-time registry only | simple, type-safe | third parties must fork/rebuild — disqualifying for this project (right answer for first-party tools, wrong here) |
| Go `plugin` `.so` | in-process speed | Linux-only, exact-toolchain ABI lock — disqualifying |
| **Tiered: (1) built-in providers in-tree, (2) generic HTTP/exec provider for zero-code wiring, (3) `hashicorp/go-plugin` gRPC for full-fidelity external plugins, (4) WASM (wazero) as sandboxed future tier** | each adopter pays only the complexity they need; language-agnostic from tier 2 up; CI-friendly (subprocess model fits one-shot runs) | protocol/contract must be versioned; more surface to document |
| WASM only | strong sandbox, polyglot | forces a wasm toolchain on every plugin author today; ecosystem still maturing |

## Decision (proposed)

**Tiered provider model behind one `FactProvider` / `PermissionProvider` port:**

1. **Built-ins** (in-tree, config-activated): forge group membership (GitLab/GitHub), OIDC/
   Keycloak group lookup, LDAP, ownership-file (CODEOWNERS-style) — covering the common cases
   with zero plugin code.
2. **HTTP / exec provider**: declare an endpoint or executable in config; assent calls it
   with a versioned JSON request and expects a versioned JSON response. Any language, no SDK.
3. **gRPC plugins** (`hashicorp/go-plugin`): for providers needing streaming, caching hooks,
   or richer lifecycle; subprocess model matches the one-shot CI execution well.
4. **WASM (wazero)** — reserved future tier for sandboxed, hot-loadable providers; recorded as
   reversible option, not built in v1.

All tiers implement the same request/response contract; the contract (not the transport) is
the versioned public API.

## Consequences

- Provider results become **facts** in PolicyInput — policies never call providers directly,
  keeping evaluation pure/deterministic and trivially testable (facts are fixtures in tests).
- Caching, timeouts, and failure semantics (fail-open vs. fail-closed per provider) must be
  spec'd; default is **fail-closed → human review**.

## Counterpoints considered

- *"Just do HTTP webhooks, skip go-plugin."* — Possible v1 simplification; tier 3 may be
  deferred if tier 2 proves sufficient in the design spike. Tracked as open question OQ-4.

## Amendment (2026-07-21, adversarial review F4/F6)

- `failure: open` is **forbidden** for any provider whose facts are referenced by a `vouch`
  rule or a risk threshold — `assent lint` fails the config. Fail-open is only legal for
  purely informational (`comment`) facts.
- Provider results carry resolution timestamps; the full resolved fact set is recorded in the
  decision report (`Pins`) for hermetic replay, and fact freshness at merge time is bounded
  per [ADR-0015 §3](0015-trust-boundaries-merge-integrity.md).
