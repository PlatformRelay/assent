# P2-E1 — Spike A: CEL residual code risk

**Problem**: ADR-0013 chose CEL-leaf trees on cel-go; acceptance needs evidence on the five
named residual risks: (1) numeric coercion YAML/HCL→CEL, (2) error UX for missing
facts/unknown fields, (3) cost limit + purity of the standard env, (4) per-leaf trace
wiring, (5) **one activation model** serving `assert` leaves *and* `{{ }}` CEL message
interpolation (ADR-0016 §2). Spike code is throwaway (lives in `hack/spikes/`, excluded
from the D-010 coverage gate) but its tests are real and runnable.
**Scope**: spike harness + report. **Non-goals**: production evaluator, envelope loader,
Rego backend (locked, D-012).
ADRs: 0013 (consequences list), 0016 §2, 0007 tri-state; OQ-11 residual.

## P2-E1-S01 — Numeric coercion decision (highest risk)

- **Goal**: decide `CrossTypeNumericComparisons` vs adapter-side normalization with an
  executable coercion table.
- **Operator input**: no.
- **Dependencies**: P1-E2-S02 (real fixture values drive the case table).
- **Definition of done**: coercion test covers at minimum: YAML `12` vs `12.0`, quoted
  `"12"`, YAML 1.1 octal-ish `010`, `yes/no` booleans, HCL number vs string, int64 overflow
  values; report records the chosen strategy and the exact cel-go env options.

Requirements:

- **REQ-P2-E1-S01-01** — Given the coercion case table, when
  `go test ./hack/spikes/cel/...` runs, then every case asserts the evaluated result (or
  typed error) under the chosen strategy, and `new >= old` comparing YAML int to float
  yields a defined, documented outcome — never a silent false.
  - Test: `hack/spikes/cel/coercion_test.go`
  - Verify: `go test ./hack/spikes/cel/ -run TestCoercion`
  - Level: L0
- **REQ-P2-E1-S01-02** — Given the results, when the report is written, then
  `docs/planning/spikes/spike-a-cel.md` contains a `## Coercion decision` section naming the
  strategy and the rejected alternative with reasons.
  - Test: `docs/planning/spikes/spike-a-cel.md`
  - Verify: `grep -q "Coercion decision" docs/planning/spikes/spike-a-cel.md`
  - Level: doc

## P2-E1-S02 — Error UX, tri-state mapping, per-leaf trace

- **Goal**: prove cel-go errors (missing fact key, unknown field, type mismatch, cost-limit
  hit) map cleanly onto ADR-0007's tri-state (`error` never takes the permissive direction,
  never the `onFail` branch) and that a per-leaf trace (leaf id, result, message) can be
  captured for `explain`/`Finding`.
- **Operator input**: no.
- **Dependencies**: P2-E1-S01 (shared env setup).
- **Definition of done**: harness walks an `all/any/not` tree, produces a trace record per
  leaf; the four error classes are distinguishable in the trace.

Requirements:

- **REQ-P2-E1-S02-01** — Given a rule tree with one failing and one erroring leaf, when the
  spike walker evaluates it, then the trace identifies which leaf failed vs errored with
  distinct machine states, and the erroring leaf reports its error class.
  - Test: `hack/spikes/cel/trace_test.go`
  - Verify: `go test ./hack/spikes/cel/ -run TestTrace`
  - Level: L0
- **REQ-P2-E1-S02-02** — Given a `vouch`-shaped (obligation-proving) predicate that errors,
  when tri-state mapping applies, then the harness asserts the obligation is NOT satisfied
  (adversarial case: an attacker-induced type error must not prove an obligation), per
  ADR-0007 amendment 1.
  - Test: `hack/spikes/cel/tristate_test.go`
  - Verify: `go test ./hack/spikes/cel/ -run TestTristate`
  - Level: L0

## P2-E1-S03 — Cost/purity + one activation model (assert + messages)

- **Goal**: measure a realistic cost budget; audit the standard env for nondeterminism
  (no time, rand, I/O); prove one activation model serves `assert` leaves and `{{ }}`
  interpolation in `message`, with unknown fields failing at *load* time (never `<no value>`,
  ADR-0016 §2).
- **Operator input**: no.
- **Dependencies**: P2-E1-S01.
- **Definition of done**: report section records the recommended cost limit with
  measurements, the purity audit result, and the interpolation compile-check design;
  ADR-0013/0016 evidence links ready for P2-E5.

Requirements:

- **REQ-P2-E1-S03-01** — Given a message template `"quota {{ facts.quota.max }} exceeded"`,
  when compiled against the declared activation, then a typo'd field
  (`facts.qota.max`) is rejected at compile time with a positioned error.
  - Test: `hack/spikes/cel/interpolation_test.go`
  - Verify: `go test ./hack/spikes/cel/ -run TestInterpolation`
  - Level: L0
- **REQ-P2-E1-S03-02** — Given the archetype predicates from P1-E2, when evaluated twice
  under the cost budget, then results are identical (determinism double-run) and all stay
  under the recommended budget; the budget and headroom are recorded in the report.
  - Test: `docs/planning/spikes/spike-a-cel.md`
  - Verify: `grep -qi "cost" docs/planning/spikes/spike-a-cel.md && go test ./hack/spikes/cel/ -run TestDeterminism`
  - Level: L0
