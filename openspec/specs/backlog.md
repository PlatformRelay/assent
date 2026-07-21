# assent backlog — phase → epic index

Spec-driven backlog per [meta-plan](../../docs/planning/meta-plan.md) and the rules in
[openspec/config.yaml](../config.yaml). Phases 1–2 carry full stories/REQs (one directory per
epic); Phases 3–5 are epic paragraphs in [later-phases.md](later-phases.md) and get full
stories during Phase 3 ("contracts first").

Design-phase epic IDs: `P1-En` / `P2-En`. Implementation epics (Phase 5): `E1..E13`.
REQ ID format: `REQ-<epic>-S<story>-<nn>` (e.g. `REQ-P1-E2-S01-01`).

## Phase 1 — Requirements harvest

| ID | Epic | Spec | Status | Needs operator | Gate contribution |
| --- | --- | --- | --- | --- | --- |
| P1-E1 | Sample corpus: generalized repos + OSS corpus (OQ-16, D-008) | [spec](p1-e1-sample-corpus/spec.md) | Ready (S02 unblocked; S01 waits on operator) | yes (repo shapes) | fixtures for the archetype gate |
| P1-E2 | Rule-archetype inventory + success metric (OQ-25) | [spec](p1-e2-archetype-inventory/spec.md) | Ready | yes (holdout adjudication) | **the Phase-1 gate**: every archetype has example change + expected decision |
| P1-E3 | Forge behaviour dossier: GitLab + GitHub (OQ-7/18/23) | [spec](p1-e3-forge-dossier/spec.md) | Ready | no | feeds ADR-0005/0017 acceptance |
| P1-E4 | Prior-art review | [spec](p1-e4-prior-art/spec.md) | Ready | no | feeds ADR acceptance round |

**Phase-1 gate**: archetype inventory reviewed; every archetype has ≥1 concrete example
change + expected decision written down (P1-E2-S02).

## Phase 2 — Spikes & ADR firming

| ID | Epic | Spec | Status | Needs operator | Gate contribution |
| --- | --- | --- | --- | --- | --- |
| P2-E1 | Spike A — CEL: coercion, error UX, cost/purity, trace, activation model | [spec](p2-e1-spike-cel/spec.md) | Blocked on P1-E2-S02 (fixture cases) | no | ADR-0013/0016 acceptance evidence |
| P2-E2 | Spike B — GitLab-in-kind vs testcontainer (OQ-6) | [spec](p2-e2-spike-e2e/spec.md) | Ready | no | ADR-0006 acceptance evidence |
| P2-E3 | Spike C — typed HTTP/exec provider contract + token isolation | [spec](p2-e3-spike-provider/spec.md) | Ready | no | ADR-0004/0017 §6 acceptance evidence |
| P2-E4 | Secure-setup adoption spike (OQ-24) | [spec](p2-e4-spike-secure-setup/spec.md) | Blocked on P1-E3-S01/S02 | yes (timed clean-room run, real repo) | north-star wording; Phase-4 prerequisite |
| P2-E5 | ADR acceptance round (0002–0017 → Accepted/Superseded) | [spec](p2-e5-adr-acceptance/spec.md) | Blocked on all P2 spikes (P2-E6 excluded) + P1-E3/E4 | yes (operator review) | **the Phase-2 gate** |
| P2-E6 | Spike D — Kubernetes CRD/CR validation feasibility (D-017 B11) | [spec](p2-e6-spike-crd/spec.md) | Ready (parallel; does **not** gate P2-E5) | no | feeds ADR-0020 + the E14 go/no-go |

**Phase-2 gate**: every Proposed ADR moved to Accepted or Superseded, each with matrix +
spike evidence (P2-E5-S02).

## Phases 3–5

Epic paragraphs (goal, ADR constraints, exit gate, story seeds) in
[later-phases.md](later-phases.md). Summary:

| Phase | Epics | Gate |
| --- | --- | --- |
| 3 — Contracts first | P3-E1 schemas + contract fixture (incl. ApprovalEvidence + named-consumer fixture) · P3-E2 versioning/compat spec · P3-E3 example migration · P3-E4 lifecycle: phase/profiles/comparison (ADR-0018) · P3-E5 publication reconciliation protocol (ADR-0019) | strict end-to-end contract fixture validates (ADR-0017 §8, D-016); new ADRs 0018/0019 accepted at the freeze review |
| 4 — Walking skeleton | P4-E1 (+ rerun-idempotence gate, D-017) | L3 skeleton green + **one real repo on live MRs** (D-012) |
| 5 — Implementation | E1–E9 active; E11/E12 **unlocked** (D-017, post-Phase-4); E14 gated on Spike D; E10/E13 **locked** (D-012) | per-epic; E7 starts alongside E1 |

Named-consumer disposition (what unlocked, what stayed locked, and why):
[docs/planning/named-consumer-compat.md](../../docs/planning/named-consumer-compat.md).

## Reading order

1. [docs/vision.md](../../docs/vision.md) → [meta-plan](../../docs/planning/meta-plan.md)
2. ADR-0017 (contract model — newest, reshapes 0003/0005/0007/0009/0010/0011/0014/0015),
   then ADR-0013, 0014, 0015, 0016
3. [open-questions.md](../../docs/planning/open-questions.md) +
   [decisions.md](../../docs/decisions/decisions.md) (D-010, D-012, D-016)
4. This index → Phase-1 epic specs → Phase-2 epic specs → [later-phases.md](later-phases.md)
