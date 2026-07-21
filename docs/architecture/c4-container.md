# C4 — Level 2: Containers / components

Hexagonal: a pure decision core, ports for everything with a side effect.

```mermaid
flowchart LR
    subgraph cli["verdict2 CLI (one static binary, one run per MR)"]
        direction TB
        subgraph inbound["Ingestion"
            ]
            forgeIn["Forge adapter (read)\nGitLab | GitHub"]
            parsers["Format adapters\nJSON · YAML · HCL/tfvars"]
            differ["Structural differ\n→ canonical ChangeSet"]
        end
        subgraph core["Pure decision core (no I/O)"]
            input["PolicyInput\n(ChangeSet + facts + MR metadata)"]
            engine["Policy engine"]
            rego["Rego frontend (OPA)"]
            yamlf["Declarative YAML frontend"]
            decision["Decision + Findings\nAPPROVE | REVIEW | BLOCK"]
        end
        subgraph providers["Provider host"]
            builtin["Built-ins: forge groups,\nOIDC/Keycloak, LDAP, owners-file"]
            httpexec["HTTP / exec providers"]
            grpc["gRPC plugins (go-plugin)"]
        end
        subgraph outbound["Publication"]
            forgeOut["Forge adapter (write)\nthreads · comments · approve · merge"]
            report["Machine-readable report\n(JSON artifact for audit/replay)"]
        end
    end

    forgeIn --> parsers --> differ --> input
    providers -- facts --> input
    rego --> engine
    yamlf --> engine
    input --> engine --> decision
    decision --> forgeOut
    decision --> report
```

## Contracts (public, versioned)

| Contract | Consumers |
| --- | --- |
| **PolicyInput** schema | policy authors (Rego + YAML), test harness |
| **Decision/Findings** schema | forge adapters, audit tooling, test harness |
| **Provider** request/response | plugin authors (HTTP, exec, gRPC) |
| **Forge port** semantics | adapter implementers; defined by the conformance suite |

## Package sketch (subject to spec phase)

```
cmd/verdict2/          CLI entrypoints: run, test, lint, render
internal/core/         engine, decision model, policy loading   (pure)
internal/change/       value tree, differ, ChangeSet            (pure)
internal/format/       json | yaml | hcl adapters
internal/policy/       rego frontend, yaml frontend
internal/provider/     provider host + built-ins
internal/forge/        port + gitlab | github adapters
internal/harness/      user-facing policy test runner
```
