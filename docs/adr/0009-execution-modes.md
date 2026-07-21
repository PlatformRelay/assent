# ADR-0009: Execution modes: CI, local/dry-run, explain, webhook service

| | |
| --- | --- |
| **Status** | Proposed |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0005 forge](0005-forge-abstraction-gitlab-first.md) · [ADR-0007 effects](0007-rule-effects-decision-aggregation.md) |

## Context

The same decision pipeline must run: in CI per MR (primary); on a developer's machine before
pushing ("what would the gate say?"); in a debugging session ("*why* did it say that?"); and —
for orgs that prefer event-driven operation over per-repo CI jobs — as a webhook receiver.
The v1 scaffold said "no long-lived service"; that hardens into: **the core is a one-shot
pipeline; the service is a thin wrapper around it**, so serve-mode support is an architecture
constraint now even if the wrapper ships later.

## Decision (proposed)

One core pipeline (ingest → classify → evaluate → aggregate → publish), four entrypoints:

| Mode | Command | Publisher | Notes |
| --- | --- | --- | --- |
| CI | `run` | real forge writes | reads MR context from CI env vars (GitLab CI first) |
| Local / dry-run | `run --dry-run` (also default when no MR context) | **recorder**: prints decision, findings, and every action it *would* take | works against local branch vs target ref; no token needed unless providers require it |
| Explain / debug | `explain` (or `run --explain`) | recorder + full trace | per-change: detected classes, routed packs/bindings, matched rules, predicate results, score arithmetic, aggregation path |
| Webhook service | `serve` | real forge writes | long-lived; subscribes to MR events, clones/checks out the branch per event (ADR-0008 §4), then runs the identical pipeline |
| Historical scan | `scan --since <date> \| --mrs <range>` | recorder only | replays past (even merged) MRs through the current policy set: backtesting a pack before enabling automerge, calibration ("what % would have automerged?"), regression checks after policy changes. Emits one JSON report per MR |
| Statistics | `stats <reports-glob>` | n/a | aggregates JSON reports (from `run` or `scan`) into automerge rate, outcome distribution, top firing rules, score histograms — flat files, **no database for now** (ADR-0012) |

Side effects are isolated behind the **Publisher** port; dry-run swaps in a recorder — the
decision core cannot tell the difference. The JSON report is emitted in every mode and is
byte-identical between a dry-run and a real run on the same input (determinism gate).

Shipping order: `run`/`--dry-run`/`explain` in v1; `serve` in v1.x once the CI path is proven
(OQ-14) — but ports and CLI structure assume it from day one.

## Consequences

- CI env parsing (which vars identify the MR) is adapter code, not core; adding GitHub
  Actions later touches only that layer + forge adapter.
- `serve` introduces state concerns (event dedup, re-evaluation on thread resolution) that the
  one-shot modes don't have; these must be spec'd before it ships, not bolted on.
- Every doc example can show the dry-run first — the adoption path starts with zero risk.

## Counterpoints considered

- *"Webhook-first like a bot framework."* — Event-driven is operationally heavier (state,
  HA, secrets custody) and most target orgs can add a CI job trivially; CI-first keeps the
  trust story simple ("it runs in *your* pipeline with *your* token").
