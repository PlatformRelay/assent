# P2-E3 — Spike C: typed HTTP/exec provider contract + token isolation

**Problem**: ADR-0017 §6 requires a typed, minimized provider protocol (declared outputs,
projection-only requests, versioned envelopes, `resolved|unavailable|invalid|expired`
states); ADR-0015 §7 requires the write token never reach a provider. v1 scope is
builtin + HTTP + exec only — gRPC/WASM are out per D-012 (this spike must NOT design them).
**Scope**: draft schemas, toy provider both transports, isolation harness. **Non-goals**:
Phase-3 schema freeze; provider caching design.
ADRs: 0004 + amendments, 0015 §7, 0017 §6; OQ-17 residual (per-fact maxAge defaults).

## P2-E3-S01 — Draft the typed contract and prove it on a toy provider

- **Goal**: draft `request.schema.json` / `response.schema.json` (declared projections,
  typed outputs with `type`, `cardinality`, `subject`, `sensitive`, `maxAge`; response
  states with `observedAt`/`expiresAt`); implement a toy group-membership provider twice —
  HTTP server and exec — both validating against the same schemas.
- **Operator input**: no.
- **Dependencies**: none (archetype facts from P1-E2 inform field choices).
- **Definition of done**: both transports return identical envelopes for identical queries;
  all four fact states are producible; the report proposes per-fact-type `maxAge` defaults
  (OQ-17 residual) for P2-E5.

Requirements:

- **REQ-P2-E3-S01-01** — Given a FactQuery for group membership, when sent over HTTP and
  exec, then both responses validate against `response.schema.json` and are byte-identical
  after canonicalization.
  - Test: `hack/spikes/provider/contract_test.go`
  - Verify: `go test ./hack/spikes/provider/ -run TestContract`
  - Level: L0
- **REQ-P2-E3-S01-02** — Given a provider that times out / returns garbage / returns a
  stale `expiresAt`, when the harness processes the response, then the fact lands in
  `unavailable` / `invalid` / `expired` respectively — distinct machine states, never a
  silently absent key, and never `resolved` (controlling facts fail closed).
  - Test: `hack/spikes/provider/states_test.go`
  - Verify: `go test ./hack/spikes/provider/ -run TestStates`
  - Level: L0

## P2-E3-S02 — Token isolation and minimization, adversarially

- **Goal**: prove the isolation invariants hold under a hostile provider: the forge write
  token never enters the provider's environment, argv, stdin payload, or HTTP request; the
  request carries only declared projections.
- **Operator input**: no.
- **Dependencies**: P2-E3-S01.
- **Definition of done**: a deliberately malicious toy provider (dumps env + full request to
  its output) is run by the harness and the assertion proves zero token/credential material
  and zero undeclared change-content reached it; report section `## Isolation evidence`.

Requirements:

- **REQ-P2-E3-S02-01** — Given the harness holds `ASSENT_FORGE_TOKEN`, when the malicious
  exec provider dumps its entire environment and stdin, then the dump contains neither the
  token value nor any variable matching `*TOKEN*`/`*SECRET*` passed through.
  - Test: `hack/spikes/provider/isolation_test.go`
  - Verify: `go test ./hack/spikes/provider/ -run TestIsolation`
  - Level: L0
- **REQ-P2-E3-S02-02** — Given a provider whose declaration requests only
  `values.pointers: [/owner]`, when the harness builds the request for a change touching
  `/owner` and `/secretRef`, then the request contains the `/owner` projection only; a
  provider requesting full old/new without the explicit trusted capability is refused at
  config-load in the harness.
  - Test: `hack/spikes/provider/minimization_test.go`
  - Verify: `go test ./hack/spikes/provider/ -run TestMinimization`
  - Level: L0
