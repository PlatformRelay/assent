# P1-E3 — Forge behaviour dossier: GitLab + GitHub

**Problem**: ADR-0005/0015/0017 acceptance and the Reconcile port design need *verified* API
mechanics, not assumptions — especially approval-eligibility evidence for `require-review`
(OQ-23), merge-result pinning options (ADR-0017 §1), and GitHub challenge parity
(OQ-7/OQ-18). Per-tier capability gaps decide where auto-merge is refused (fail closed).
**Scope**: two dossier documents with API-cited evidence (endpoint, version, tier).
**Non-goals**: adapter code; GitHub implementation (E10, locked per D-012 — the dossier only
keeps the seam honest).
ADRs: 0005, 0009 amendment, 0015 §2/§4/§5, 0017 §1/§3/§7; OQ-7, OQ-18, OQ-23.

## P1-E3-S01 — GitLab dossier: threads, approvals, merge preconditions, tiers

- **Goal**: document, with API endpoint citations per claim: resolvable discussions +
  all-discussions-resolved merge gate; approvals + self-approval limits; bot identity via
  project token; `merge?sha=` CAS semantics; merge trains/queues availability per tier
  (merge-result pinning, ADR-0017 §1); protected-pipeline topologies (ADR-0015 §4);
  "merge when pipeline succeeds" arming (ADR-0009 amendment).
- **Operator input**: no.
- **Dependencies**: none.
- **Definition of done**: `docs/planning/forge-dossier-gitlab.md` complete; every capability
  row states Free/Premium/Ultimate availability and the fail-closed consequence of absence.

Requirements:

- **REQ-P1-E3-S01-01** — Given the dossier, when a Reconcile precondition from ADR-0017 §7
  (source SHA, target SHA, merge-result digest) is listed, then the dossier names the exact
  GitLab mechanism that can enforce or verify it — or records `capability gap → no deferred
  auto-merge on this tier`.
  - Test: `docs/planning/forge-dossier-gitlab.md`
  - Verify: `grep -q "merge-result" docs/planning/forge-dossier-gitlab.md && grep -qi "capability gap" docs/planning/forge-dossier-gitlab.md`
  - Level: doc
- **REQ-P1-E3-S01-02** — Given ADR-0015 §2, when the SHA-guard is documented, then the
  dossier records the observed behaviour of `PUT /merge_requests/:iid/merge?sha=` when HEAD
  moved (verified against a live gitlab.com test project, response codes captured), including
  the adversarial case: push after evaluation, then attempt the pinned merge → must fail.
  - Test: `docs/planning/forge-dossier-gitlab.md`
  - Verify: `grep -q "sha=" docs/planning/forge-dossier-gitlab.md`
  - Level: doc

## P1-E3-S02 — Approval-eligibility evidence for `require-review` (OQ-23)

- **Goal**: establish per GitLab tier how the adapter *proves* an eligible approval:
  approval-rules API, CODEOWNERS-sourced eligible principals, who counts, whether the bot's
  own approval is excludable — feeding ADR-0017 §3's "typed eligible principals".
- **Operator input**: no.
- **Dependencies**: P1-E3-S01.
- **Definition of done**: OQ-23 has a leading answer; dossier section states, per tier,
  either the evidence chain or `capability gap → require-review never satisfiable → no
  auto-merge for archetypes needing it`.

Requirements:

- **REQ-P1-E3-S02-01** — Given a tier, when `require-review` evidence is documented, then
  the dossier lists the API calls returning (a) required approval rules, (b) eligible
  approvers, (c) actual approvals with identities — and the adversarial case: an approval by
  the MR author or by assent's own bot identity must be distinguishable and excluded.
  - Test: `docs/planning/forge-dossier-gitlab.md`
  - Verify: `grep -qi "eligible" docs/planning/forge-dossier-gitlab.md`
  - Level: doc

## P1-E3-S03 — GitHub dossier: parity mapping for the designed seam

- **Goal**: answer OQ-7/OQ-18 on paper: can `REQUEST_CHANGES` + required-conversation-
  resolution + SHA-pinned auto-merge reproduce the arm-and-wait flow; who dismisses the bot
  review; base-ref workflow trust for forks (ADR-0015 §4); GitHub merge-queue as the
  merge-result-pinning mechanism.
- **Operator input**: no.
- **Dependencies**: P1-E3-S01 (structure reuse).
- **Definition of done**: `docs/planning/forge-dossier-github.md` complete; explicitly marked
  "informs the Forge-port seam only — E10 locked per D-012"; OQ-7/OQ-18 rows updated.

Requirements:

- **REQ-P1-E3-S03-01** — Given the GitLab dossier's capability rows, when the GitHub dossier
  is authored, then every row has a GitHub mapping, a parity verdict
  (`parity | partial | gap`), and — for `require-review` — the CODEOWNERS/required-reviewers
  evidence chain equivalent.
  - Test: `docs/planning/forge-dossier-github.md`
  - Verify: `grep -qi "parity" docs/planning/forge-dossier-github.md`
  - Level: doc
