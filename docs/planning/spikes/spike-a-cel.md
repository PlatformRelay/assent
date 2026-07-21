# Spike A — CEL residual code risk (P2-E1)

**Status**: complete (harness in `hack/spikes/cel/`, throwaway; tests real).  
**cel-go**: `github.com/google/cel-go v0.29.2`  
**ADRs**: 0013 consequences, 0016 §2, 0007 tri-state amendment 1.  
**Fixture sources**: `examples/archetypes/` (P1-E2) — bounded-change partitions 12→16,
quota 24; ownership `entry.owner` / `facts.author.groups`.

## Coercion decision

**Chosen strategy: `cel.CrossTypeNumericComparisons(true)` + thin adapter guards —
preserve decoded numeric identity; never silently parse strings as numbers.**

Env options (see `hack/spikes/cel/activation.go`):

- `cel.CrossTypeNumericComparisons(true)` — so `new >= old` with YAML int `12` and float
  `16.0` (or typed `int`/`double` decls) evaluates to a defined bool, never a silent false
  and never a mysterious compile failure when both sides are numeric.
- `cel.Variable("old"|"new", cel.DynType)` for change scalars; typed `cel.Facts` via
  `ext.NativeTypes` + `ext.ParseStructTag("cel")` for facts.
- Adapter helpers `AsCELNumber` / `AsCELBool` reject quoted numerics, bool/string confusion,
  and values outside int64 before they reach the interpreter.

**Rejected alternative: adapter-side normalization that rewrites every number to
`float64` (or every scalar through `strconv`).** Reasons: loses integer identity (equality
and overflow behaviour change); masks author mistakes (`"12"` quoted, HCL string labels);
pushes policy bugs into silent float equality. Cross-type compare at the CEL layer is the
smaller, better-scoped fix for the real archetype hazard (int partitions vs float facts).

### Executable table highlights (`TestCoercion`)

| Case | Outcome under chosen strategy |
| --- | --- |
| YAML `12` vs `16.0`, `new >= old` | **true** (CrossType) |
| Archetype int→int 12→16 | **true** |
| Quoted `"12"` vs int | CEL **`no such overload`** (not silent false) |
| YAML `010` (octal-ish) | yaml.v3 decodes as **8** — author hazard; document decimal-only |
| YAML `yes`/`no` | stay **strings**; CEL `== true` is **silent false** → **`AsCELBool` rejects** |
| HCL number vs int | compares; HCL string vs int → overload error |
| float beyond int64 | **`AsCELNumber` → overflow error** |

## Error UX / tri-state / per-leaf trace

Walker (`Walk`) over `all`/`any`/`not` emits a `LeafTrace` per leaf: id, `pass|fail|error`,
message, and `ErrorClass` when errored.

| Class | Example |
| --- | --- |
| `missing_fact` | map key absent (`no such key`) |
| `unknown_field` | `facts.qota.*` at compile |
| `type_mismatch` | `no such overload` / `type conversion error` |
| `cost_limit` | `CostLimit` exceeded |

**ADR-0007**: `VouchSatisfied` is true only for `pass`. Adversarial type error on an
obligation leaf (`TestTristate`) does **not** prove the obligation; errors never take
`onFail`.

**Residual hazard (documented):** dyn `==` and `in` return **false** (not error) on
cross-type values. Production must keep facts/entry on NativeTypes (or equivalent schema
validation) so mistakes become compile/type errors — the spike uses conversion/`startsWith`
where dyn equality would lie.

## Cost / purity / one activation model

### Cost budget

| Predicate (archetype-shaped) | Measured cost |
| --- | --- |
| bounded-change `new >= old && new <= facts.quota.max_partitions` | 8 |
| ownership `entry.owner in facts.author.groups` | 6 |
| allow-listed `path in […]` | 14 |
| **max** | **14** |

**Recommended budget: `1000`** (`CostBudget` in `activation.go`).  
**Headroom: 986** (≈70× max archetype cost). Enforced via `cel.CostLimit(1000)` +
`cel.CostTracking(nil)`. Double-run (`TestDeterminism`) yields identical bool + cost.

### Purity audit (standard env)

Registered: variables + NativeTypes for `Facts` only. **Not** registered: `ext.Strings`
time helpers beyond baseline, custom I/O, `rand`, wall-clock bindings. Standard cel-go
macros (`map`/`filter`/`all`/`exists`) are pure. Spike assertions use no wall-clock.

### Message interpolation (ADR-0016 §2)

Same `NewEnv()` activation compiles `{{ … }}` slots (`CompileMessage`).

- `"quota {{ facts.quota.max_partitions }} exceeded"` → OK  
- typo `facts.qota.max_partitions` → **compile error** with `message template:line:col`
  (never `<no value>`)

## Evidence links for P2-E5

- ADR-0013 residual list → this report + `hack/spikes/cel/*_test.go`
- ADR-0016 §2 one activation / load-time fields → `TestInterpolation`
- ADR-0007 tri-state → `TestTristate` / `TestTrace`
