# P1-E1 — Sample corpus: generalized sample repos + OSS corpus

**Problem**: Phase-1 harvest and all later fixtures need representative governed-repo
content. Nothing private may be committed (D-002); operator shapes are *re-created*, never
copied (D-008); public corpora (OQ-16) give demo/test targets with zero sanitization risk.
**Scope**: content under `examples/repos/`, corpus selection record, sanitization check.
**Non-goals**: generator scripts for e2e seeding (E7), policy packs for the samples (Phase 3+).
ADRs: D-002, D-008, OQ-16; layout per `examples/repos/README.md`.

## P1-E1-S01 — Generalize operator repo shapes into committed samples

- **Goal**: the three planned shapes (`topic-registry/` YAML, `service-catalog/` JSON,
  `infra-vars/` tfvars) exist as small, realistic, fully generic sample repos.
- **Operator input**: **yes** — 2–3 real repo-shape descriptions (structure, fields,
  governance semantics), delivered as notes; never files from private repos.
- **Dependencies**: none (operator input gates completion, not start — seed from vision).
- **Definition of done**: three sample dirs committed; sanitization check runnable and green;
  each README names which archetypes the shape exercises.

Requirements:

- **REQ-P1-E1-S01-01** — Given the operator's shape notes, when samples are authored, then
  each of `examples/repos/topic-registry/`, `examples/repos/service-catalog/`,
  `examples/repos/infra-vars/` contains ≥3 governed entry files with owner-bearing entries,
  env split (`prod`/`dev` paths), and a `README.md` stating the governed workflow and
  exercised archetypes.
  - Test: `examples/repos/topic-registry/README.md`
  - Verify: `test -f examples/repos/topic-registry/README.md && test -f examples/repos/service-catalog/README.md && test -f examples/repos/infra-vars/README.md`
  - Level: doc
- **REQ-P1-E1-S01-02** — Given D-002, when any commit adds sample content, then
  `hack/check-sanitization.sh` scans the tree against a workspace-local denylist
  (path from `ASSENT_SANITIZE_DENYLIST`, never committed) plus built-in generic patterns
  (internal-looking hostnames, `*.corp`, employee-ID shapes) and exits non-zero on a hit.
  Adversarial case: a denylisted term inside a base64-decodable YAML value is still caught
  (script decodes obvious base64 values before matching).
  - Test: `hack/check-sanitization.sh`
  - Verify: `bash hack/check-sanitization.sh`
  - Level: doc

## P1-E1-S02 — Select and pin the open-source corpus (OQ-16)

- **Goal**: 2–3 public corpora chosen from the candidate table, pinned by commit SHA, with a
  per-corpus mapping to archetypes — resolving OQ-16.
- **Operator input**: no (recommendation authored; operator ratifies in P2-E5).
- **Dependencies**: none.
- **Definition of done**: `examples/repos/corpus.md` committed; OQ-16 row updated to
  "leading answer" with link.

Requirements:

- **REQ-P1-E1-S02-01** — Given the candidates in `examples/repos/README.md`, when selection
  is made, then `examples/repos/corpus.md` records for each chosen corpus: repo URL, pinned
  commit SHA, license, relevant paths, file format, and which archetypes its change history
  exercises; and records for each *rejected* candidate a one-line reason.
  - Test: `examples/repos/corpus.md`
  - Verify: `grep -q "commit" examples/repos/corpus.md && grep -qi "license" examples/repos/corpus.md`
  - Level: doc
- **REQ-P1-E1-S02-02** — Given a chosen corpus, when its shape is needed as a fixture, then a
  ≤10-file excerpt is vendored under `examples/repos/corpus/<name>/` with an attribution
  header (source repo, SHA, license) in a `NOTICE` file — no license-incompatible content.
  - Test: `examples/repos/corpus/`
  - Verify: `test -d examples/repos/corpus && find examples/repos/corpus -name NOTICE | grep -q NOTICE`
  - Level: doc
