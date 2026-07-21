# Prior-art review

Meta-plan Phase 1.4 — steal shamelessly, differentiate deliberately. For each tool: what it got
right/wrong *for our use case* — a deterministic merge gate for governed self-service repos
([vision.md](../vision.md)). Every section answers the same five questions — **change-model
approach**, **policy surface**, **merge-integrity approach (TOCTOU handling)**, **testing story
for rules**, **adoption friction** — and ends with `Lessons for assent:`.

Scope guard: the closing implications table maps each lesson to an ADR or OQ it confirms, amends,
or challenges. Implications may *feed* the P2-E5 acceptance round; they do not re-litigate
decided ADRs (D-006 stays decided).

## 1. OPA / conftest

- **Change model** — none. conftest evaluates the *current state* of structured files (YAML,
  JSON, HCL, Dockerfile, …) as a single `input` document; `--combine` merges multiple files for
  cross-file rules ([conftest.dev](https://www.conftest.dev/)). There is no `old`/`new` pair —
  a policy that cares about *what changed* (bounded increase, deletion, ownership of the touched
  entry) must reconstruct the delta itself, outside the tool.
- **Policy surface** — Rego. `deny`/`violation`/`warn` rules in a `policy/` directory, external
  data via `--data`, policy distribution via OCI bundles. Full expressive power, full learning
  curve.
- **Merge integrity (TOCTOU)** — none, by design. conftest is a pass/fail CI step; it never
  approves or merges, so it inherits whatever the forge's branch protection does. If the target
  branch moves after the check ran, nothing re-evaluates unless branch protection demands
  up-to-date branches.
- **Testing story** — first-class and the best in this list: `conftest verify` runs Rego unit
  tests (`*_test.rego`, `with input as` overrides, `parse_config`/`parse_config_file` helpers)
  ([docs](https://www.conftest.dev/)); plain `opa test` adds coverage reporting. Policy tests are
  a normal part of the workflow, including as pre-commit hooks.
- **Adoption friction** — Rego. Platform engineers who are not programmers routinely stall on
  it; that reputation is the reason declarative front-ends (Kyverno, and our tier 1) exist at
  all. Otherwise adoption is a single static binary — the model assent copies (ADR-0001).

Lessons for assent: state-assertion tools cannot express diff-shaped policy — this is the gap
the canonical change model (ADR-0003) exists to fill, and why ADR-0002 feeds Rego a computed
`PolicyInput` instead of raw files. `conftest verify` is the bar for our adopter harness
(ADR-0006 L1, ADR-0014): policy tests must be one command, zero infrastructure.

## 2. Kyverno (incl. the ValidatingPolicy/CEL direction)

- **Change model** — the strongest in this list, courtesy of Kubernetes admission review:
  policies see `object` *and* `oldObject`, so update rules can compare states natively. Since
  1.14 the `ValidatingPolicy` type is CEL-first (a superset of Kubernetes
  `ValidatingAdmissionPolicy`), stable in 1.18, and can evaluate arbitrary JSON/YAML payloads via
  `spec.evaluation.mode: JSON`; the legacy JMESPath-based `ClusterPolicy` is deprecated
  ([policy-types overview](https://kyverno.io/docs/policy-types/overview/),
  [1.14 announcement](https://kyverno.io/blog/2025/04/25/announcing-kyverno-release-1.14/)).
- **Policy surface** — declarative YAML CRDs; in the modern types, CEL expressions with
  per-expression `message`s, `matchConditions`, fine-grained exceptions, extended CEL libraries
  (HTTP/resource lookups) ([ValidatingPolicy docs](https://kyverno.io/docs/policy-types/validating-policy/)).
- **Merge integrity (TOCTOU)** — solved by architecture, not effort: admission control sits at
  the *single synchronous choke point* (the API server) and judges the exact object about to be
  persisted. There is no gap between evaluation and commit. A merge gate has no such choke point
  unless the forge provides one — which is exactly why merge queues exist (§4) and why our
  preconditions must carry what admission gets for free.
- **Testing story** — `kyverno test` runs policies against resource fixtures with expected
  results (unit level); Chainsaw covers e2e. Testing is documented as a product surface, not an
  afterthought.
- **Adoption friction** — Kubernetes-shaped: CRDs, controllers, cluster context. For non-K8s
  payloads the CLI works but the ecosystem's centre of gravity is admission. The
  ClusterPolicy → ValidatingPolicy migration also shows the cost of evolving an authored policy
  API underneath users: a whole ecosystem of written policies goes legacy.
- **Syntax note** — kyverno-json (the standalone JSON engine) was evaluated and dropped in
  ADR-0013: dormant ~18 months, bus factor ≈ 1, and its assertion trees are structurally wrong
  for two-state (`old`/`new`) payloads.

Lessons for assent: CEL-in-YAML with per-expression messages *is* contemporary Kyverno style —
ADR-0013's hybrid tree is riding the ecosystem's own trajectory, and `object`/`oldObject` is the
direct precedent for our `entry`/`oldEntry` predicate scope. The deprecation churn is the
cautionary tale behind freezing authored schemas strictly and versioning them (ADR-0017 §9)
before anyone writes policies against them.

## 3. Mergify

- **Change model** — PR metadata attributes: files touched, labels, author, branch, check
  states, review states — matched by a condition grammar (`files~=^prod/`,
  `check-success=ci`). No content-level model: it can see *that* `prod/topics.yaml` changed,
  never *which field* changed by how much.
- **Policy surface** — declarative `.mergify.yml`: `pull_request_rules` (conditions → actions)
  and `queue_rules` with separate `queue_conditions`/`merge_conditions`
  ([file format](https://docs.mergify.com/configuration/file-format/)). Notably, the
  configuration used to evaluate a PR comes from the base branch, so a PR cannot rewrite the
  rules that gate it.
- **Merge integrity (TOCTOU)** — a real merge queue: PRs are updated against the latest base and
  re-checked before merge; speculative checks build temporary draft PRs of cumulative merges
  (base + #1, base + #1 + #2, …) and `merge_conditions` are evaluated against that temporary
  merge result; a push to a queued PR dequeues it
  ([performance docs](https://docs.mergify.com/merge-queue/performance/)). Instructive wart: the
  `skip_intermediate_results` option lets a PR merge even when its own speculative check failed,
  as long as a later batch containing it passed — throughput pressure producing an integrity
  escape hatch.
- **Testing story** — schema validation of the config, but no fixture-based decision simulation
  an adopter can run offline against "here is a hypothetical PR". You learn what your conditions
  do by watching them act on live PRs.
- **Adoption friction** — a SaaS GitHub App holding write/merge permission on your repos: for
  many platform teams that is a procurement and trust problem before it is a technical one. The
  condition grammar itself is easy.

Lessons for assent: metadata-level conditions are where policy expressiveness dies — file-glob
gating is precisely the bespoke-bot ceiling the canonical change model (ADR-0003) breaks
through. Config-loaded-from-base-branch independently confirms ADR-0015 §1, and the
`skip_intermediate_results` wart is a warning to never grow throughput knobs that trade away
merge-result integrity (ADR-0017 §1 keeps fail-closed as the only degradation).

## 4. Bors / merge queues (incl. GitHub merge queue)

- **Change model** — deliberately none. A PR is an opaque set of commits; the only question is
  "does CI pass on the exact result of merging it".
- **Policy surface** — a reviewer command (`bors r+`) plus repo config in bors-ng; in GitHub's
  native merge queue, branch protection / rulesets: required checks + "require merge queue"
  ([GitHub docs](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/managing-a-merge-queue)).
  Authorization to enqueue *is* the policy; everything else is CI.
- **Merge integrity (TOCTOU)** — the gold standard, solved *by construction* (Graydon Hoare's
  "Not Rocket Science Rule": never ship anything that wasn't tested as the thing you ship).
  bors-ng merges the batch into a `staging` branch, runs CI there, and fast-forwards main to the
  bit-for-bit tested commit ([bors-ng README](https://github.com/bors-ng/bors-ng)). GitHub's
  merge queue creates `gh-readonly-queue/*` branches containing base + queued PRs, dispatches
  `merge_group` check runs, and merges only when those pass. The evaluated thing and the merged
  thing cannot diverge.
- **Testing story** — n/a: there are no rules to test; the test suite is the policy.
- **Adoption friction** — CI must be retrofitted to trigger on the queue's synthetic refs
  (`merge_group` event / queue branch patterns) — a classic silent-failure setup step. Queue
  latency on busy repos. And the ecosystem lesson: bors-ng is deprecated and its public instance
  retired *because* the platform absorbed the feature
  ([bors newsletter #76](https://bors.tech/newsletter/2023/04/30/tmib-76/)).

Lessons for assent: this is the strongest external confirmation of ADR-0017 §1 — pinning the
source SHA alone (`merge?sha=`) is exactly the mistake merge queues exist to prevent; a decision
is only valid for a pinned *merge result*, so deferred auto-merge must ride a forge queue/train
or re-evaluate on target movement. And bors-ng's fate says: never compete with forge-native
merge mechanics — compose with them (the decision layer is our product, the queue is theirs);
on GitHub, running assent as a required `merge_group` check is the natural way to get
merge-result re-evaluation for free (feeds OQ-18).

## 5. Renovate automerge

- **Change model** — the best proof that a *semantic* change model is what makes automerge
  policy writable: Renovate knows the package, the old and new version, the update type
  (patch/minor/major), the manager — and `packageRules` match on those fields, not on diff
  text ([config options](https://docs.renovatebot.com/configuration-options/)). It is
  domain-specific (dependency bumps only); assent generalizes the same idea to arbitrary
  structured files.
- **Policy surface** — declarative `renovate.json`: `packageRules` matchers + `automerge`,
  `minimumReleaseAge` (let the ecosystem soak a release before trusting it),
  `internalChecksFilter` — risk appetite expressed as data.
- **Merge integrity (TOCTOU)** — automerge only when the branch is up-to-date with base *and*
  green; if the base moved, Renovate rebases and re-tests on the next run
  ([automerge key concepts](https://docs.renovatebot.com/key-concepts/automerge/)). Two
  documented warts: on busy repos the up-to-date requirement means automerge is *starved* by
  perpetual rebasing; and with `platformAutomerge`, if branch protection requires no status
  checks, "GitHub might automerge PRs with failing tests" — arming a forge-native mechanism
  whose preconditions nobody verified is a fail-open trap. `ignoreTests: true` exists as an
  explicit (discouraged) escape hatch.
- **Testing story** — none for the policy itself: `renovate-config-validator` checks schema,
  but there is no "given this update, what would you do?" fixture harness; adopters discover
  rule behaviour in production or in dry-run logs.
- **Adoption friction** — the trust ramp is cultural, not technical: docs recommend starting
  with patch-level devDependencies and widening gradually, backed by real test coverage. Noise
  (PR-created + PR-merged notifications) pushed them to `automergeType=branch` — silent merges
  without a PR — trading auditability for calm.

Lessons for assent: Renovate is the strongest independent validation of ADR-0003 (policy
against a semantic change model, not diffs) — and its platform-automerge footgun is the exact
scenario `assent doctor` exists to prevent: verify the forge-side gate configuration before
arming anything (ADR-0015 §4, ADR-0017 §9). The staged-trust rollout culture maps onto our
`off`/`observe`/`enforce` phases (OQ-21) and `scan` backtesting (ADR-0009); the silent-merge
trade-off warns us to keep the decision record mandatory even when the UX is quiet (ADR-0016).

## 6. Prow / Tide

- **Change model** — none at content level: Tide selects PRs by GitHub search criteria — labels
  (`lgtm`, `approved`, absence of `do-not-merge/*`), branch, check states
  ([Tide docs](https://docs.prow.k8s.io/docs/components/core/tide/)). Content-aware judgment is
  delegated to humans, routed by OWNERS files: path-scoped approver lists in-repo decide *who*
  may `/approve` what — mechanized human authorization rather than mechanized judgment.
- **Policy surface** — central Prow `config.yaml` (tide queries, presubmit jobs, merge methods)
  owned by the platform team, plus per-repo OWNERS. Split-brain by design: merge criteria are
  central, ownership is local.
- **Merge integrity (TOCTOU)** — a long-lived reconciler, not a one-shot: Tide syncs every ~1
  minute and "ensures that PRs are tested against the most recent base branch commit before they
  are allowed to merge", retesting automatically whenever any other merge makes results stale;
  batches are tested as a unit ([Tide docs](https://docs.prow.k8s.io/docs/components/core/tide/),
  [PR author's guide](https://docs.prow.k8s.io/docs/components/core/tide/pr-authors/)). Also a
  stated *security* posture: with Tide, humans are not allowed to click merge at all
  ([Prow at scale](https://docs.prow.k8s.io/docs/scaling/)).
- **Testing story** — Prow config changes are code-reviewed and there are config linters, but
  there is no adopter-facing harness simulating "would this PR merge?"; the label semantics are
  simple enough that the community relies on the `tide` status context to explain pool
  membership.
- **Adoption friction** — enormous: a Kubernetes cluster running a dozen components, a bot
  token, ghproxy for API rate limits. It amortizes across an org with hundreds of repos and is
  wildly oversized for one self-service repo — the gap assent's one-CI-job install targets.
- **Return-of-experience** — re-test-before-merge is only *possible* because Tide is a service
  that can observe base movement and act later. A one-shot CI job structurally cannot promise
  that.

Lessons for assent: Tide proves both halves of ADR-0009's arming model — continuous
re-test-before-merge requires a reconciler (our `serve` tier, OQ-14), and therefore a one-shot
run must never arm what it cannot revoke (ADR-0017 §4): what one-shot mode *can* do honestly is
pin the merge result and delegate enforcement to forge-native gates. OWNERS is the canonical
prior art for path-scoped ownership facts (ADR-0004 providers) and for forge-proven approval as
the only real authorization (`require-review`, ADR-0017 §3 / OQ-23).

## 7. danger.js

- **Change model** — PR metadata plus raw git surface handed to imperative code:
  `danger.git.modified_files`, diffs as strings, `danger.github.pr`
  ([danger.systems/js](https://danger.systems/js/)). Any structural understanding (parse the
  YAML, compare fields) is the Dangerfile author's problem, rebuilt per repo.
- **Policy surface** — an imperative Dangerfile in JavaScript/TypeScript with `fail`/`warn`/
  `message`/`markdown` outputs. A full programming language: maximal power, zero declarative
  legibility — the rule *is* its implementation.
- **Merge integrity (TOCTOU)** — none: danger is advisory. It posts a comment / fails a check
  and never approves or merges, so it carries no TOCTOU exposure — and provides no automation
  either. (Peril, the hosted event-driven variant, is archived.)
- **Testing story** — technically possible (a Dangerfile is code; the docs show mocking
  `danger` in Jest), practically absent: real-world Dangerfiles are untested glue scripts. There
  is no fixture-decision harness; nothing *requires* tests.
- **Adoption friction** — trivially easy to start, which is the trap: the five-line Dangerfile
  grows into the untested bespoke bot vision.md describes — logic invisible to the governed,
  dying with its author. Fork-PR token handling is a recurring pain (writes need a token the
  fork's CI must not hold — cf. ADR-0015 §8's execution-authority matrix).

Lessons for assent: danger.js is the closest ancestor of the bespoke-bot pattern we are built to
replace — it validates the core bet that gate logic must be declarative, diff-shaped, and
test-required (ADR-0002, ADR-0003, ADR-0006 L1) rather than imperative and optional-everything;
its comment-first UX (explain findings *in* the review surface) survives into our presentation
layer (ADR-0012).

## 8. The conda-forge / Homebrew bot-automerge precedent

- **Change model** — the narrowest possible: a known bot authors PRs of a known shape
  (version + hash + URL bumps). conda-forge automerges PRs "from the `regro-cf-autotick-bot`
  with `[bot-automerge]` in the title, all statuses passing, and the feedstock allows automerge"
  (opt-in `bot.automerge` in `conda-forge.yml`)
  ([infrastructure docs](https://conda-forge.org/docs/maintainer/infrastructure/)); Homebrew's
  BrewTestBot auto-merges bottle PRs once a maintainer with write access approved and CI is
  green ([BrewTestBot docs](https://docs.brew.sh/BrewTestBot-For-Maintainers)). Trust is
  anchored on *author identity + change class*, not on inspecting the diff.
- **Policy surface** — flags and labels: per-repo opt-in config, `automerge` /
  `automerge-skip` / `new formula` labels, title markers. The actual logic lives in centrally
  maintained bot code (conda-forge-webservices / brew `pr-publish`), invisible to the governed
  repo.
- **Merge integrity (TOCTOU)** — forge required-status checks plus low contention; conda-forge
  serializes racy CI jobs with a turnstyle action rather than evaluating merge results. Works
  because feedstock/formula repos have almost no concurrent human traffic — a luxury a generic
  gate cannot assume.
- **Testing story** — none at the policy level: the "policy" is bot source code, tested (if at
  all) as software by the central team; a feedstock maintainer cannot simulate "would this PR
  automerge?".
- **Adoption friction** — near zero for the individual repo (one config key / one label) —
  because a central platform team operates the whole apparatus across thousands of homogeneous
  repos. The model does not transfer to heterogeneous self-service repos without generalizing
  the bot into a product — which is the assent thesis.

Lessons for assent: the largest packaging communities independently converged on
bot-automerge-for-routine-changes — demand for the product category is proven at scale, and
opt-in per-repo enablement plus staged trust is the adoption pattern to copy (`.assent/`
config, OQ-21 phases). But "trust the author + green CI" is a special case assent must
generalize: author identity becomes just another typed fact, and safety comes from named
obligations over the change itself (ADR-0017 §2), not from hardcoding who the bot is.

## Consolidated implications

Effect legend: **confirms** (independent evidence for a decided/proposed design), **amends**
(suggests sharpening wording or adding a case, without reopening the decision), **challenges**
(pressure the design must answer — routed to the P2-E5 acceptance round, not silently adopted),
**feeds** (input to an open question).

| # | Prior art | Implication | ADR / OQ | Effect |
| --- | --- | --- | --- | --- |
| 1 | Bors staging branch; GitHub `merge_group` | A decision is valid only for a pinned **merge result**; source-SHA CAS alone is the exact bug merge queues exist to fix | ADR-0017 §1 | confirms |
| 2 | GitHub merge queue mechanics | On GitHub, run assent as a required `merge_group` check → merge-result re-evaluation comes free from the forge | ADR-0017 §1 · OQ-18 | feeds |
| 3 | Prow/Tide re-test-before-merge via 1-minute reconciler loop | Genuine re-test/re-evaluate at merge time requires a service; one-shot mode must arm only what stays forge-enforced — Tide is the existence proof for both the arming constraints and the `serve` tier | ADR-0009 (amendment) · ADR-0017 §4 · OQ-14 | confirms |
| 4 | Tide: "humans are not allowed to merge" as security posture | Bot-approval-as-the-gate should be stated as a *feature* with defense-in-depth options, exactly as ADR-0015 §5 does | ADR-0015 §5 | confirms |
| 5 | Mergify evaluates config from the base branch | Policy loaded from target ref, never the judged branch | ADR-0015 §1 | confirms |
| 6 | Mergify `skip_intermediate_results`; Renovate `ignoreTests` | Adopters will demand integrity escape hatches for throughput/noise; prior art grew them and regrets are documented — assent's fail-closed-only degradation needs an explicit acceptance-round defense of the UX cost | ADR-0017 §1 (counterpoint) · feeds P2-E5 | challenges |
| 7 | Renovate platform-automerge with unconfigured branch protection merges failing PRs | Arming forge-native automerge without verifying forge gate config is fail-open; `doctor` as arming precondition is mandatory, not nice-to-have | ADR-0015 §4 · ADR-0017 §9 | confirms |
| 8 | Renovate `packageRules` on semantic update facts | Policy over a semantic change model (not diff text) is what makes automerge rules writable — the domain-specific proof of our generic bet | ADR-0003 | confirms |
| 9 | Renovate automerge starvation on busy repos (perpetual rebase) | Deferred automerge on high-traffic repos may rarely fire; adoption docs and the secure-setup spike should measure/expect this | OQ-24 | feeds |
| 10 | Kyverno ValidatingPolicy: CEL + per-expression messages | ADR-0013's hybrid (condition tree, CEL leaves, per-leaf message) rides the current mainstream of declarative policy; `object`/`oldObject` precedes `entry`/`oldEntry` | ADR-0013 | confirms |
| 11 | Kyverno ClusterPolicy → ValidatingPolicy deprecation churn | Authored-schema evolution under users' feet burns an ecosystem; freeze strict, versioned schemas before packs exist | ADR-0017 §9 · meta-plan Phase 3 gate | confirms |
| 12 | conftest `verify` / `opa test` | Fixture-based policy unit tests as a one-command product surface; the bar for `assent test` | ADR-0006 L1 · ADR-0014 | confirms |
| 13 | conftest state-only input (no old/new) | State-assertion languages cannot express diff policy; Rego tier must consume the computed change model, not raw files | ADR-0002 · ADR-0003 | confirms |
| 14 | Prow OWNERS path-scoped approvers | Ownership-by-path as data, and forge-proven approval as the only real authorization — direct precedent for permission providers and `require-review` | ADR-0004 · ADR-0017 §3 · OQ-23 | confirms |
| 15 | danger.js untested imperative Dangerfiles; archived Peril | The bespoke-bot failure mode is real and common; declarative + test-required is the differentiator, and fork-PR token handling needs the execution-authority matrix | ADR-0002 · ADR-0015 §8 | confirms |
| 16 | bors-ng deprecated once the forge shipped merge queues | Never compete with forge-native merge mechanics; the durable product layer is the *decision*, composing with the forge's queue/auto-merge primitives | ADR-0017 §1 · vision non-goals | amends (sharpens the composition stance) |
| 17 | conda-forge/Homebrew: opt-in flags, staged trust, central bot code | Category demand proven at scale; copy per-repo opt-in + staged rollout; generalize "trusted author" into typed facts + named obligations instead of hardcoded identities | ADR-0017 §2 · OQ-21 · OQ-25 | feeds |
