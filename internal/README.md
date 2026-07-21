# internal/

Intentionally empty until specs exist (meta-plan Phase 3/4) — packages follow contracts, not
the other way around. Planned sketch (see
[C4 containers](../docs/architecture/c4-container.md)):

```
internal/core/       engine, decision model, policy loading   (pure — no I/O)
internal/change/     value tree, structural differ, ChangeSet (pure)
internal/format/     json | yaml | hcl adapters
internal/policy/     rego frontend, declarative yaml frontend
internal/provider/   provider host + built-ins
internal/forge/      port + gitlab | github adapters
internal/harness/    adopter-facing policy test runner
```

Architecture rule (to be enforced with go-arch-lint once code lands): `core` and `change`
import neither adapters nor anything doing I/O.
