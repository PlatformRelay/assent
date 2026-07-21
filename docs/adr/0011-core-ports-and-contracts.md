# ADR-0011: Core Go ports and public contracts (draft shapes)

| | |
| --- | --- |
| **Status** | Accepted (partial: Reconcile + schemas-are-API per ADR-0017 §1/§7; P2-E5) |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0003](0003-canonical-change-model.md) · [ADR-0004](0004-plugin-architecture.md) · [ADR-0005](0005-forge-abstraction-gitlab-first.md) · [ADR-0007](0007-rule-effects-decision-aggregation.md) · [ADR-0009](0009-execution-modes.md) |

## Context

Contracts before code (meta-plan Phase 3). These sketches exist to *get a feeling* for the
seams and to be attacked in review; exact shapes freeze in Phase 3 with fixtures. Everything
here is draft.

## Data contracts (serialized, versioned — the real public API)

```go
// --- change model (ADR-0003, ADR-0008) ---
type ChangeSet struct {
	Files   []FileEvent // added | deleted | renamed | modified | opaque
	Changes []Change
}
type Change struct {
	File        string   // repo-relative path
	Path        string   // RFC-6901-style pointer within the file
	Kind        Kind     // add | modify | delete
	Old, New    Value    // typed scalars/trees; nil per Kind
	Classes     []string // set by classifier
	Environment string   // set by classifier
}

// --- policy input: everything a predicate may see (pure) ---
type PolicyInput struct {
	Change  *Change           // the matched change (change scope)
	Changes []Change          // all changes in this class slice
	Branch  BranchState       // branch scope only: parsed trees at head SHA (lazy)
	Facts   map[string]any    // provider results, keyed by provider name
	MR      MergeRequestMeta  // author, source/target branch, labels, forge
}

// --- decision (ADR-0007) ---
type Finding struct {
	Rule    string
	Effect  Effect // comment | challenge | block | vouch | score
	Points  int
	Paths   []string
	Message string
}
type Decision struct {
	Outcome   Outcome // APPROVE | REVIEW | BLOCK
	Findings  []Finding
	Score     int
	Threshold int
	Trace     Trace // classes, bindings, per-rule eval — powers `explain`
	Pins      Pins  // head SHA, policy SHA, tool version — replayability (OQ-9)
}
```

## Ports (Go interfaces at the hexagon boundary)

```go
// Format adapters (ADR-0003): one per file type, registry-selected by extension.
type FormatAdapter interface {
	Match(path string) bool
	Parse(data []byte) (Value, error)
}

// Predicate backends (ADR-0002): assert-tree/CEL and rego implement this.
type PredicateBackend interface {
	Compile(rule RuleSpec) (Predicate, error) // compile-once, at policy load
}
type Predicate interface {
	Eval(in PolicyInput) (Result, error) // pure; no I/O possible by construction
}

// Providers (ADR-0004): permissions are just facts; one port, four transports
// (builtin | http | exec | grpc). Resolved BEFORE evaluation; results become
// PolicyInput.Facts — predicates never call out.
type Provider interface {
	Name() string
	Resolve(ctx context.Context, q FactQuery) (map[string]any, error)
}
// FactQuery carries MR author, touched classes/paths — enough for a permission
// service to answer "which groups / which owned entries" in one round-trip.

// Forge (ADR-0005): read side.
type Forge interface {
	MergeRequest(ctx context.Context, ref MRRef) (MergeRequestMeta, error)
	EnsureCheckout(ctx context.Context, ref MRRef, dir string) (BranchInfo, error)
	Threads(ctx context.Context, ref MRRef) ([]Thread, error) // resolution state
}

// Publisher (ADR-0007/0009): write side; one method per effect, plus verdict.
// Dry-run/explain swap in a recorder implementation — core can't tell.
type Publisher interface {
	Comment(ctx context.Context, f Finding) error
	OpenThread(ctx context.Context, f Finding) error // resolvable
	Approve(ctx context.Context, d Decision) error
	Deny(ctx context.Context, d Decision) error
	Merge(ctx context.Context, d Decision) error
}
```

## Invariants

- `internal/core` + `internal/change` import no port implementations (arch-lint enforced).
- `Predicate.Eval` is pure: facts pre-resolved, branch state pre-parsed (lazy but memoized),
  no clock, no randomness. This is what makes golden tests and replay trivial.
- Every contract change goes through an openspec change proposal; serialized forms carry
  `apiVersion`.

## Counterpoints considered

- *"Providers as two interfaces (PermissionProvider / FactProvider)."* — Collapsed into one:
  a permission check *is* a fact ("author's groups", "owned entries"). Fewer seams, and the
  per-company reimplementation story is one interface with four transport options.
- *"Let rules call providers on demand."* — Rejected: kills purity, caching, and dry-run
  fidelity; pre-resolution with declared provider deps keeps evaluation a pure function.

## Amendment (2026-07-21, adversarial review F3/F9/F11)

- **SHA-guarded writes (F3):** `Approve(ctx, d Decision, sha Pin)` and
  `Merge(ctx, d Decision, sha Pin)` — adapters must use the forge's compare-and-swap
  (GitLab `merge?sha=`, GitHub merge `sha`); on mismatch they fail closed. Conformance-suite
  case, per ADR-0015 §2.
- **Hermetic pins (F9):** `Pins` additionally records the **full resolved fact set** and
  per-provider resolution timestamps. Replay of a historical decision re-uses pinned facts —
  never re-resolves. `scan` resolves facts at scan time unless a fact snapshot is supplied;
  its report must carry a `facts: live` caveat flag that `stats` surfaces next to any
  backtest percentage.
- **Per-change predicate binding (F11):** a rule's predicate is evaluated **once per matched
  change**, with `old/new/path/kind/file/entry/oldEntry` bound to that change (scope table in
  the ADR-0013 appendix). A `vouch` covers exactly the changes whose predicate returned true;
  false or error leaves that change uncovered (tri-state per ADR-0007 amendment). `entry` /
  `oldEntry` (containing entry at head/base) are added to the PolicyInput contract.

## Amendment 2 (2026-07-21, second review P1-4/P1-5/P2-11)

- **Content-keyed facts (P1-4):** `FactQuery` carries the class-sliced ChangeSet (paths,
  old/new values, classes, environment) — not just author+paths — and `config.yaml` provider
  entries may declare **key extractors** (JSON-pointer expressions over changed entries,
  e.g. `extract: { costCenter: "/metadata/costCenter" }`) whose extracted values arrive in
  the query. Without this, any "referenced X must exist" provider is unimplementable.
- **Publisher lifecycle ops (P1-5):** the port gains `UpsertComment` (idempotent, marker-
  keyed), and `SyncThreads` (open new / resolve stale / **reopen or re-post when a resolved
  thread's underlying value changed**) driven by the finding-lifecycle state machine in
  ADR-0012 amendment 2. `Comment`/`OpenThread` alone cannot express idempotent re-runs.
- **Positions (P2-11):** `Change` carries file+line/column spans (ADR-0003 amendment 2) so
  findings can anchor inline threads; `Finding` gains an optional `Anchor`.
