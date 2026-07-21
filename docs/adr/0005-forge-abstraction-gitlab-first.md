# ADR-0005: Forge abstraction: GitLab first, GitHub second

| | |
| --- | --- |
| **Status** | Proposed |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0006 testing](0006-testing-strategy.md) |

## Context

The tool must act like a reviewer on both major forges: post findings as **resolvable review
threads**, comment, approve/deny, and auto-merge. The primitives differ: GitLab has MR
discussions with per-thread resolution and an "all discussions resolved" merge gate; GitHub
has PR reviews (`REQUEST_CHANGES`/`APPROVE`) and review-thread resolution with different
semantics. Self-approval restrictions, bot identities, and merge APIs also differ.

## Options

| Option | Pros | Cons |
| --- | --- | --- |
| **Forge-neutral `Forge` port with capability flags; GitLab adapter first, GitHub second; e2e conformance suite runs against both** | core stays platform-free; conformance suite *defines* the port semantics; capability flags make gaps explicit instead of leaky | port design must resist "GitLab-shaped" bias — mitigated by writing the GitHub mapping into the spec from day one |
| GitLab-only v1, abstract later | fastest MVP | retrofitting an abstraction under a shipped behaviour contract is the classic trap |
| Lowest common denominator | simple port | wastes the strongest feature of each forge (e.g. GitLab resolvable-thread merge gate) |

## Decision (proposed)

**Forge port designed for both from day one; GitLab adapter implemented first.** The port is
specified in behavioural terms ("publish findings such that merging is blocked until each is
acknowledged/resolved") and each adapter maps that to native primitives — GitLab: blocking
discussions + all-resolved merge gate; GitHub: `REQUEST_CHANGES` review + thread resolution.
Where a forge cannot express a behaviour, the adapter declares a capability gap and the engine
falls back (documented, not silent). A single **conformance test suite** runs against both
adapters (kind-hosted GitLab; GitHub test org) and is the executable definition of the port.

## Consequences

- CI entrypoints stay thin: a GitLab CI template and (later) a GitHub Action wrap the same CLI.
- Auth models per forge (project token / GitHub App) live entirely in the adapter.

## Counterpoints considered

- *"Abstractions before the second implementation are guesses."* — True in general; here the
  second implementation's API is fully known and stable, so the mapping can be spec'd (not
  guessed) up front, and the conformance suite catches drift.
