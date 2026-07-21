# C4 — Level 2: Containers / components

Hexagonal: a pure decision core, ports for everything with a side effect.

```mermaid
flowchart LR
    subgraph cli["assent CLI (one static binary, one run per MR)"]
        direction TB
        subgraph inbound["Ingestion"]
            forgeIn["Forge adapter (read)\nGitLab | GitHub"]
            policyLoad["Policy loader\n(.assent/ from TARGET ref — ADR-0015)"]
            parsers["Format adapters\nJSON · YAML · HCL/tfvars\n(+ positions, resource limits)"]
            differ["Structural differ + classifier\n→ canonical ChangeSet (classes, env)"]
        end
        subgraph core["Pure decision core (no I/O)"]
            input["PolicyInput\n(ChangeSet + facts + MR metadata)"]
            engine["Policy engine\n(envelope: match · effects · onFail · points)"]
            assertB["assert backend\n(CEL-leaf trees, cel-go — ADR-0013)"]
            regoB["rego backend\n(OPA, capability-sandboxed)"]
            decision["Decision + Findings + Trace + Pins\nAPPROVE | REVIEW | BLOCK"]
        end
        subgraph providers["Provider host (no forge write token)"]
            builtin["Built-ins: forge groups,\nOIDC/Keycloak, LDAP, owners-file"]
            httpexec["HTTP / exec providers\n(digest-pinned)"]
            grpc["gRPC / WASM tiers"]
        end
        subgraph outbound["Publication"]
            renderer["Renderer (escaped, redacted,\nfinding lifecycle state machine)"]
            forgeOut["Forge adapter (write)\nthreads · comments · SHA-guarded\napprove/merge · auto-merge arm"]
            report["JSON report artifact\n(hermetic pins: SHA, policy, facts)"]
        end
    end

    forgeIn --> parsers --> differ --> input
    policyLoad --> engine
    providers -- facts --> input
    assertB --> engine
    regoB --> engine
    input --> engine --> decision
    decision --> renderer --> forgeOut
    decision --> report
```

## Contracts (public, versioned)

| Contract | Consumers |
| --- | --- |
| **PolicyInput** schema (incl. predicate scope) | policy authors, test harness |
| **Decision/Findings/Pins** schema | forge adapters, audit tooling, `stats`, test harness |
| **Provider** request/response (content-keyed FactQuery) | plugin authors (HTTP, exec, gRPC, WASM) |
| **Forge port** semantics (SHA-guarded writes, thread lifecycle) | adapter implementers; defined by the conformance suite |
| **Test fixture format** (ADR-0014) | adopters |

## Package sketch (subject to spec phase)

```
cmd/assent/          CLI entrypoints: run, test, lint, explain, scan, stats, doctor, init
internal/core/       engine, decision model, aggregation                  (pure)
internal/change/     value tree, differ, classifier, ChangeSet            (pure)
internal/format/     json | yaml | hcl adapters (positions, limits)
internal/policy/     envelope loader (target-ref), assert backend, rego backend
internal/provider/   provider host + built-ins (token-isolated)
internal/forge/      port + gitlab | github adapters (SHA-guarded writes)
internal/render/     renderer, finding lifecycle, redaction
internal/harness/    adopter-facing policy test runner
```
