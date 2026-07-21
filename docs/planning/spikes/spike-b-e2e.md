# Spike B — GitLab-in-kind vs GitLab CE testcontainer (P2-E2, resolves OQ-6)

**Question** (ADR-0006, OQ-6): what is the CI default for L3 e2e — GitLab CE as a plain
Docker container ("testcontainer" profile) or GitLab hosted in a kind cluster?
**Method**: `hack/spikes/e2e/boot-testcontainer.sh` and `hack/spikes/e2e/boot-kind.sh`,
3 cold-boot runs each on the same host; then the product-surface smoke
(`hack/spikes/e2e/smoke.sh`) on the winner. Raw CSVs kept out of git (oss-playbook avoid-#5).

## Setup

- Host: macOS arm64 (Apple Silicon), Docker Desktop VM with 6 CPU / ~11.7 GiB RAM, kind v0.33.
- Image: `gitlab/gitlab-ce:19.2.0-ce.0` (multi-arch; native arm64 — no emulation needed).
  Pinned; pre-pulled outside the timer so both profiles compare cached-image cold boot.
- Slimmed `GITLAB_OMNIBUS_CONFIG` **identical in both profiles** (affects the RAM numbers):
  prometheus stack, container registry, KAS, and outbound mail disabled; puma in single mode
  (`worker_processes 0`); sidekiq concurrency 5; monitoring whitelist opened for the
  host-side readiness probe.
- Readiness = HTTP 200 from `/-/readiness`; RSS sampled after a 60 s settle.
- kind profile: existing `hack/kind/kind-config.yaml` scaffold (NodePort 30080 → host 8929),
  CE-in-pod Deployment (no Helm) with the same image and Omnibus config, image side-loaded
  via archive (clock paused — host-cache artifact). `boot_seconds` **includes kind cluster
  creation**, since that is the profile's real per-run cost in CI.

## Measurements

| Profile | Runs | Boot p50 (s) | Boot range (s) | Steady-state RSS (MB) | Flakes |
| --- | --- | --- | --- | --- | --- |
| testcontainer (docker run) | 3 | 96 | 90–96 | ~2440 (2424–2461) | 0 |
| kind-hosted (kind + CE pod) | 3 | 126 | 125–127 | ~3140 (2894–3302) | 0 |

Notes:

- Boot is far below the 3–8 min folklore for GitLab CE — the slimmed Omnibus config and
  native arm64 image help a lot. Numbers will be slower on shared CI runners.
- The kind profile pays a constant overhead (cluster create ~40 s + pod scheduling) on top
  of the same GitLab boot, plus the node's own memory footprint (not included in the RSS
  column, which is the GitLab container cgroup only).
- One tooling flake was hit and fixed during the spike: `kind load docker-image` fails
  against Docker's containerd image store (multi-platform manifest, single-platform blobs);
  the script now exports a single-platform archive and uses `kind load image-archive`.
  Counted as tooling friction, not a boot flake (readiness itself never flaked).

## Product-surface smoke (winner)

`hack/spikes/e2e/smoke.sh` on the **testcontainer** profile, exercising the Reconcile-port
primitives (ADR-0011, ADR-0015 §2): create project (root PAT minted at runtime) → seed from
`examples/repos/topic-registry/topics/prod/` → branch + change commit → open MR → post
resolvable discussion → resolve it → approve with a **project access token** (Maintainer
bot) → adversarial case → SHA-pinned merge. Exits 0 with all assertions passing; repeatable
(deletes and recreates the sample project each run).

**Adversarial SHA-guard case (REQ-P2-E2-S02-02, ADR-0015 §2 evidence)**: after approval, a
new commit is pushed to the source branch and the merge is attempted with the **stale** head
SHA via `PUT .../merge?sha=<stale>`. GitLab rejects it with **HTTP 409** and body
`SHA does not match HEAD of source branch: <fresh>`; the merge then succeeds with the fresh
SHA. Transient **HTTP 405** can appear while `detailed_merge_status=checking` immediately
after the push — the smoke waits for mergeability to settle before asserting the 409.

## Decision

**CI default: the GitLab CE testcontainer profile** (`docker run`, no cluster).
Rationale:

- Same GitLab boot cost dominates both profiles; kind adds ~30 s constant overhead
  (cluster create + pod scheduling) and an extra moving part (cluster lifecycle, image
  side-loading, NodePort plumbing) with zero benefit for a single-instance CI job.
- The testcontainer profile is self-contained per CI run, maps directly onto
  services/`docker run` in any CI, and its failure surface is one container.
- RAM: testcontainer ~2.4 GB vs kind ~3.1 GB for the GitLab container **plus** the kind
  node itself — the testcontainer fits standard 8 GB CI runners with headroom.

**kind stays for local/demo** (per the OQ-6 note): `hack/kind/` remains the long-lived
local instance that doubles as the demo environment and the E7 conformance-suite host,
where cluster overhead is paid once, not per run.

Follow-ups: E7 turns the smoke into the real conformance suite; E9 wires the
testcontainer profile into CI. If CI runners prove slower/flakier than this host, revisit
with runner-native measurements before E9 lands.
