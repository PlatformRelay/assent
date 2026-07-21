# ADR-0001: Implementation language: Go, single static binary

| | |
| --- | --- |
| **Status** | Accepted |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0002 policy frontends](0002-policy-frontends-rego-declarative.md) · [ADR-0004 plugins](0004-plugin-architecture.md) |

## Context

The tool runs **inside CI pipelines** (GitLab CI, later GitHub Actions): it must start fast,
ship as one artifact, and have no runtime dependencies. It must parse JSON/YAML/HCL, evaluate
Rego, talk to forge APIs, and host **third-party plugins**. The open question raised during
inception: *is Go the right language given the plugin requirement?* — Go's native `plugin`
package is famously weak (Linux-only, exact-toolchain ABI lock).

## Options

| Option | Pros | Cons |
| --- | --- | --- |
| **Go** | OPA/Rego engine is a native Go library; `wazero` (pure-Go WASM runtime) and `hashicorp/go-plugin` (gRPC subprocess) give two solid *out-of-process/sandboxed* plugin paths; `hashicorp/hcl` is the reference HCL parser; first-class GitLab/GitHub clients; single static binary, instant CI startup; k8s-ecosystem contributors feel at home | in-process dynamic loading is impractical (`plugin` pkg); generics-era but still verbose |
| Rust | performance, WASM story | Rego evaluation via regorus (less battle-tested than OPA); HCL parsing ecosystem thin; slower contribution ramp for the target audience |
| TypeScript/Node | easy contribution | needs a runtime in every CI image; OPA only via WASM bridge; weak HCL support; slow cold start |
| Python | ubiquitous | same runtime problem; OPA only out-of-process; packaging pain in CI |

## Decision

**Go.** The decisive observation: the plugin requirement does *not* imply in-process dynamic
loading. Extensibility is delivered by (a) **policy-as-data** — Rego/YAML policies are loaded
at runtime, no rebuild — and (b) **out-of-process or sandboxed plugins** for permission/fact
providers (see [ADR-0004](0004-plugin-architecture.md)), where Go's ecosystem
(`hashicorp/go-plugin`, `wazero`, plain HTTP webhooks) is strong. Meanwhile Go is the *only*
language where the three hardest dependencies — OPA, HCL, forge clients — are all first-party,
mature, native libraries, and the deployment model (one static binary in a CI job) is exactly
Go's sweet spot.

## Consequences

- One `CGO_ENABLED=0` binary + a small container image; CI integration is a one-liner.
- Plugin authors are *not* forced into Go: WASM and gRPC plugin protocols are language-agnostic.
- We accept that "drop a `.so` in a folder"-style plugins will never exist.

## Counterpoints considered

- *"A scripting language would make community contribution easier."* — The community extends
  the tool through **policies** (Rego/YAML), not through core code; the core being Go does not
  gate rule authors.
- *"Rust for correctness."* — The correctness-critical part is the decision engine, which is
  covered by golden/property tests regardless of language; ecosystem fit wins.
