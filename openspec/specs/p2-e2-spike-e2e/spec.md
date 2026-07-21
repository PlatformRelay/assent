# P2-E2 — Spike B: GitLab-in-kind vs testcontainer (OQ-6)

**Problem**: ADR-0006 leaves the CI e2e default open. Decide with measurements, then prove
the winner can actually host the product's riskiest surface (threads/approve/merge).
**Scope**: boot + measure both profiles; API smoke on the winner. **Non-goals**: the real
conformance suite (E7), CI pipeline wiring (E9).
ADRs: 0006, 0005; OQ-6.

## P2-E2-S01 — Boot and measure both profiles

- **Goal**: scripted boot of GitLab CE as (a) testcontainer and (b) kind-hosted instance;
  ≥3 runs each; measure cold boot time, steady-state RAM, and flake count.
- **Operator input**: no.
- **Dependencies**: none.
- **Definition of done**: `hack/spikes/e2e/boot-testcontainer.sh` and
  `hack/spikes/e2e/boot-kind.sh` are idempotent and tear down cleanly; measurements table
  committed to the report (raw CSVs stay out of git per oss-playbook avoid-#5).

Requirements:

- **REQ-P2-E2-S01-01** — Given a machine with Docker + kind, when either boot script runs,
  then it exits 0 with GitLab answering `/-/readiness`, and prints a single machine-readable
  result line `RESULT boot_seconds=<n> rss_mb=<n>`.
  - Test: `hack/spikes/e2e/boot-testcontainer.sh`
  - Verify: `bash -n hack/spikes/e2e/boot-testcontainer.sh && bash -n hack/spikes/e2e/boot-kind.sh`
  - Level: doc
- **REQ-P2-E2-S01-02** — Given ≥3 runs per profile, when the report is written, then
  `docs/planning/spikes/spike-b-e2e.md` contains a measurements table with rows for both
  profiles (boot p50, RAM, flakes) and a `## Decision` section naming the CI default and
  the local/demo role of the other (kind stays for demo either way, per OQ-6 note).
  - Test: `docs/planning/spikes/spike-b-e2e.md`
  - Verify: `grep -q "testcontainer" docs/planning/spikes/spike-b-e2e.md && grep -q "## Decision" docs/planning/spikes/spike-b-e2e.md`
  - Level: doc

## P2-E2-S02 — Product-surface smoke on the winner

- **Goal**: on the chosen profile: seed one generated sample repo (from `examples/repos/`),
  open an MR via API, post a resolvable thread, resolve it, approve with a project token,
  and merge with `merge?sha=` — the exact primitives the Reconcile port needs.
- **Operator input**: no.
- **Dependencies**: P2-E2-S01, P1-E1-S01 (sample content).
- **Definition of done**: smoke test is repeatable (`task` target or script), and the
  SHA-guard negative case is exercised (push after approve → pinned merge fails). Findings
  feed the GitLab dossier if behaviour differs from P1-E3 documentation.

Requirements:

- **REQ-P2-E2-S02-01** — Given a booted instance, when the smoke script runs, then it
  completes seed → MR → thread → resolve → approve → SHA-pinned merge and exits 0.
  - Test: `hack/spikes/e2e/smoke.sh`
  - Verify: `bash -n hack/spikes/e2e/smoke.sh`
  - Level: L3
- **REQ-P2-E2-S02-02** — Given the merged flow, when the adversarial step runs (new push
  after approval, then merge with the stale SHA), then the merge is rejected by the forge
  and the script asserts the rejection — evidence for ADR-0015 §2 acceptance.
  - Test: `hack/spikes/e2e/smoke.sh`
  - Verify: `grep -q "sha" hack/spikes/e2e/smoke.sh`
  - Level: L3
