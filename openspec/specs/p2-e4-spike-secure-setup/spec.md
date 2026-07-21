# P2-E4 — Secure-setup adoption spike (OQ-24)

**Problem**: ADR-0017 §9 (roast P1-8): the "under one hour" north star is unproven, and the
*secure* topology (ADR-0015 §4 protected pipeline, least-privilege token, forge merge gates)
is exactly the part adopters will skip if it is slow. Decide the ONE supported v1
GitLab-tier + topology and time it clean-room.
**Scope**: topology decision + timed run + doctor checklist draft. **Non-goals**: `doctor`
implementation (Phase 4/5); GitHub topology (locked).
ADRs: 0015 §4/§5, 0017 §9, 0009 amendment (forge merge-gate settings); OQ-24.

## P2-E4-S01 — Choose the ONE supported v1 topology

- **Goal**: from the P1-E3 dossier, select the single supported combination of GitLab tier,
  pipeline-trust model (compliance pipeline / instance template / MR-approved), token type +
  scopes, and required project settings (all-threads-resolved gate, approval rules on
  `.assent/**`); write the setup walkthrough draft + `doctor` checklist (typed
  capability/precondition report fields).
- **Operator input**: no.
- **Dependencies**: P1-E3-S01, P1-E3-S02.
- **Definition of done**: `docs/planning/spikes/spike-secure-setup.md` sections `## Supported
  topology`, `## Explicitly unsupported topologies` (author-editable `.gitlab-ci.yml` with a
  privileged token named as insecure, per ADR-0015 §4), `## Doctor checklist`.

Requirements:

- **REQ-P2-E4-S01-01** — Given the dossier's tier matrix, when the topology is chosen, then
  the doc records every precondition as a checklist row with: forge setting, API call that
  verifies it, and failure consequence (warn vs refuse-to-arm).
  - Test: `docs/planning/spikes/spike-secure-setup.md`
  - Verify: `grep -q "Supported topology" docs/planning/spikes/spike-secure-setup.md && grep -qi "unsupported" docs/planning/spikes/spike-secure-setup.md`
  - Level: doc

## P2-E4-S02 — Timed clean-room run on a real repo

- **Goal**: a person who did not write the walkthrough executes it on a real repository
  (personal/demo self-service repo qualifies per D-012) with a stopwatch, through: project
  settings, token creation, protected pipeline inclusion, sample policy copy, first dry-run
  (mocked decision output is acceptable pre-implementation — the *setup* is what is timed).
- **Operator input**: **yes** — operator (or a recruited tester) performs the run; operator
  provides the real repo.
- **Dependencies**: P2-E4-S01.
- **Definition of done**: timing log committed (per-step timestamps); OQ-24 answered;
  north-star wording confirmed or amended via a decisions.md entry; friction items filed as
  walkthrough fixes.

Requirements:

- **REQ-P2-E4-S02-01** — Given the walkthrough and a stopwatch, when the clean-room run
  completes, then the report contains a per-step timing table, the total, and a
  `## North-star verdict` section stating whether <1h holds for the supported topology —
  and if not, which wording change or setup automation is required.
  - Test: `docs/planning/spikes/spike-secure-setup.md`
  - Verify: `grep -q "North-star verdict" docs/planning/spikes/spike-secure-setup.md`
  - Level: doc
