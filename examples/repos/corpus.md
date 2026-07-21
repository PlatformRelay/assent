# Open-source corpus selection (OQ-16)

Resolves OQ-16: which public repos join the demo/test corpus. Selection made from the
candidate table in [`README.md`](README.md); the operator ratifies in P2-E5. Excerpts
(≤10 files each, with per-corpus `NOTICE` attribution) are vendored under
[`corpus/`](corpus/). All pins taken 2026-07-21 from each repo's default branch.

## Chosen corpora

### 1. kubernetes/org — live governed self-service repo (YAML)

| | |
| --- | --- |
| Repo | <https://github.com/kubernetes/org> |
| Pinned commit | `8c064d2064181a40fd59c41220f60e2e358b3dd7` |
| License | Apache-2.0 |
| Relevant paths | `config/<org>/org.yaml` (org admins/members), `config/restrictions.yaml`, `config/<org>/<team>/teams.yaml` |
| File format | YAML, multi-entry membership/config files |
| Excerpt | [`corpus/kubernetes-org/`](corpus/kubernetes-org/) |

**Archetypes its change history exercises**: ownership (people PR *themselves* into
member lists; approval is gated on org/team ownership), allow-listed fields (membership
edits are routine; org-setting edits are not), no destruction (removals reviewed),
schema validity (Prow-validated config). This is the canonical *live* instance of the
exact workflow assent gates — real MR history of self-service membership changes.

### 2. JulieOps (kafka-ops/julie) descriptor files — topic-registry shape (YAML)

| | |
| --- | --- |
| Repo | <https://github.com/kafka-ops/julie> |
| Pinned commit | `e75d005e994d237642cdb566ebc6ff943d02762b` |
| License | MIT |
| Relevant paths | `example/descriptor*.yaml`, `example/plans.yaml` |
| File format | YAML topology descriptors (projects → topics with config, consumers/producers as principals) |
| Excerpt | [`corpus/julieops/`](corpus/julieops/) |

**Archetypes the format exercises**: ownership (per-project principals), bounded change
(`num_partitions`, `replication.factor`, retention configs), no destruction (topic
removal), schema validity (`dataType`/schema references), environment split (context/
project hierarchy). A public, richer sibling of our `topic-registry/` sample. Note: the
repo is dormant (last push 2024-06) — acceptable for a *fixture corpus* (we consume the
format, not the code; the dependency-health bar in GUIDELINES applies to code deps).

### 3. octoDNS zone configs — DNS record registry shape (YAML)

| | |
| --- | --- |
| Repo | <https://github.com/octodns/octodns> |
| Pinned commit | `6ad5c2529c4633ed52a3100aba056f5e0fad0ea5` |
| License | MIT |
| Relevant paths | `tests/config/*.yaml` (zone data, top-level config, per-zone thresholds) |
| File format | YAML zone files (record name → type/ttl/value entries) |
| Excerpt | [`corpus/octodns/`](corpus/octodns/) |

**Archetypes the format exercises**: bounded change (TTL bands, value edits), no
destruction (record deletion — octoDNS itself ships per-zone change thresholds, i.e.
prior art for our risk-points idea), allow-listed fields, opaque-change fallback
(provider-specific record types). Caveat recorded honestly: the in-repo files are the
project's *test fixtures*, not a live governed registry — the *format* and the many
real-world octoDNS-managed zone repos are what we target.

## Rejected candidates

| Candidate | Reason (one line) |
| --- | --- |
| devshawn/kafka-gitops | same topic-registry shape as JulieOps but a flatter format and unmaintained since 2023 — one Kafka corpus is enough |
| Backstage `catalog-info.yaml` | catalog entries are scattered across a huge monorepo and adopter repos; the repo itself is not a governed registry to replay |
| GitHub safe-settings | good shape (org settings as YAML) but the governed content lives in each adopter's admin repo, not in the public project repo — nothing to pin |
| conda-forge feedstocks / Homebrew bumps | recipe/version bumps are code-like, not structured-config entries; kept as *prior art* for bot-automerge (Phase 1.4), not corpus |
