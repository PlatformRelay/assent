# Spike C (P2-E3) — typed HTTP/exec provider contract + token isolation

Status: complete. Code: `hack/spikes/provider/` (throwaway; tests are real).
Spec: `openspec/specs/p2-e3-spike-provider/spec.md` · ADR-0004 (+amendments), ADR-0015 §7,
ADR-0017 §6 · OQ-17 residual. Scope guard honored: HTTP + exec only, no gRPC/WASM design (D-012).

## Contract summary

Two schemas, one envelope pair, transport-agnostic (`request.schema.json`, `response.schema.json`):

- **FactQuery** (`provider.assent.dev/v1alpha1`): `queryId`, host-pinned `asOf` (RFC 3339),
  `subject {kind, id}`, requested `outputs[]`, and `projections` — the *only* change content a
  provider ever sees. `projections.values[]` entries are `{pointer, old, new}` with RFC 6901
  pointers. `deadlineMs` is a hint; the host enforces its own timeout.
- **FactResponse**: echoes `queryId`; `facts[]` has exactly one entry per requested output.
  Each fact carries its **echoed declaration** (`type`, `cardinality`, `subject`, `sensitive`,
  `maxAge`) so the host can cross-check against config, a `state` in
  `resolved | unavailable | invalid | expired`, `observedAt`, and — required iff `resolved` —
  `value` + `expiresAt` (schema-enforced via `if/then`). Non-resolved facts must not carry a
  value and carry an operator-readable `reason` instead.

Proven properties (all in `go test ./hack/spikes/provider/`):

- **Transport parity** (`TestContract`): the same toy group-membership logic served over an
  HTTP server and an exec binary (query on stdin, response on stdout) yields responses that
  both validate against `response.schema.json` and are **byte-identical after
  canonicalization** (sorted keys, compact, number text preserved). All timestamps derive from
  the host-pinned `asOf` — a provider needs no wall clock, which is what makes parity testable
  and replay hermetic.
- **Fail-closed states** (`TestStates`): the host-side `ResolveFacts` classifier produces
  exactly one fact per requested output on *every* path. Timeout → `unavailable`; garbage or
  schema-invalid or `queryId`-mismatched or omitted-output responses → `invalid`; a response
  claiming `resolved` with `expiresAt <= asOf` is rewritten host-side to `expired` (value
  dropped). Distinct machine states, never a silently absent key, never `resolved` on a
  failure path — a controlling fact is fail-closed by construction.
- **Minimization** (`TestMinimization`): the request builder intersects the provider's declared
  `values.pointers` with what the change actually touched; undeclared content (`/secretRef`
  and its values) never enters the serialized request. `fullContent` without the explicit
  `trusted-full-content` capability is refused at config load, before any query exists.

## Isolation evidence

`TestIsolation` (`hack/spikes/provider/isolation_test.go`) runs a **deliberately malicious
exec provider** (`maliciousexec/`) that exfiltrates everything it can observe — its entire
environment plus its full stdin — to stdout, under a harness that holds
`ASSENT_FORGE_TOKEN` (a canary value) and `CI_JOB_SECRET`, with an operator config that
additionally tries to pass through `UPSTREAM_TOKEN` and `LDAP_SECRET`.

Mechanism: the exec transport (`CallExec`) never inherits the host environment. `ScrubEnv`
builds the child env **from scratch** — `PATH` plus explicitly configured entries, and even
configured entries are refused when their name matches `(?i)TOKEN|SECRET`.

Asserted on the actual dump produced by the hostile provider:

- neither the forge token **value** nor any other canary secret value appears;
- no variable **name** matching `*TOKEN*`/`*SECRET*` appears (inherited or configured);
- sanity checks confirm the dump is real: the declared non-secret var (`PROVIDER_MODE`) and
  the stdin payload (`queryId`) did reach the provider.

Combined with `TestMinimization`, this demonstrates ADR-0015 §7 end to end for the exec tier:
zero credential material and zero undeclared change content reach a hostile provider. (The
HTTP tier has no env/argv surface; its request body is the same minimized `FactQuery`.)

Residual risks for the real implementation (not spike-scope): argv is unused here — the real
host must also keep credentials out of provider argv; deny-listing env names is
defense-in-depth only — the primary control is *never inheriting* the host env; exec binaries
still need digest-pinning (ADR-0015 §7) which this spike did not exercise.

## Proposed per-fact-type `maxAge` defaults (OQ-17 residual → input for P2-E5)

`maxAge` is an arming precondition (ADR-0017 §4), not advisory. Provider declarations may
shorten a default, never exceed the global `facts.max_age` cap (24h, ADR-0015 §3).

| Fact type (declaration) | Example | Default `maxAge` | Rationale |
| --- | --- | --- | --- |
| `principal` / membership & eligibility sets | group membership, approval eligibility | **1h** | authorization-bearing; revocation must propagate before a deferred merge fires |
| `boolean` authorization gates | "author is owner" | **1h** | same blast radius as principal facts |
| `string`/`integer` registry lookups | cost center exists, service registered | **24h** | slow-changing reference data; global cap applies |
| any fact declared `sensitive: true` | secret-adjacent metadata | **15m** | short-lived by policy; also subject to redaction (ADR-0012) |

Two consequences worth freezing in P2-E5: (a) a fact whose `expiresAt` lands before a
deferred auto-merge's horizon must block arming (ADR-0017 §4) rather than merely expire; (b)
defaults belong in the *host*, keyed by declaration `type` + `sensitive`, so a lazy provider
declaration inherits safe values instead of no bound.
