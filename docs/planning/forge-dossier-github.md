# Forge behaviour dossier — GitHub parity mapping (P1-E3-S03)

> **Scope guard: this document informs the Forge-port seam only — E10 locked per D-012.**
> Nothing here schedules GitHub adapter work; it exists so the Forge port (ADR-0005) is not
> designed GitLab-shaped. Answers OQ-7 and OQ-18 on paper.

Every capability row from [forge-dossier-gitlab.md](forge-dossier-gitlab.md) §1 gets a GitHub
mapping and a parity verdict: `parity` (behaviourally equivalent primitive exists), `partial`
(equivalent outcome reachable with caveats the port must expose), `gap` (no equivalent —
capability flag off, engine falls back per ADR-0005). Docs snapshot: docs.github.com,
REST `2022-11-28`, 2026-07-21.

Availability note (GitHub's analogue of GitLab tiers): branch protection / rulesets and
required reviews are free on **public** repos; on **private** repos they require a paid plan;
merge queue requires an organization repo (public, or private/internal on Enterprise Cloud).
Rows below flag availability only where it changes the verdict; the E10-era dossier refresh
must re-verify plan gating per feature.

## 1. Capability parity matrix

| GitLab row | GitHub mapping (method + endpoint) | Verdict | Notes / fail-closed consequence |
| --- | --- | --- | --- |
| C1 post resolvable finding thread | review comments: `POST /repos/{o}/{r}/pulls/{n}/comments` or a review with inline comments `POST /repos/{o}/{r}/pulls/{n}/reviews` — [Reviews API](https://docs.github.com/en/rest/pulls/reviews?apiVersion=2022-11-28#create-a-review-for-a-pull-request); threads readable via GraphQL `PullRequest.reviewThreads` (`isResolved`, `viewerCanResolve`) — [GraphQL Pulls](https://docs.github.com/en/graphql/reference/pulls) | parity | resolution state lives in GraphQL only; REST cannot read/write `isResolved` |
| C2 resolve / reopen a thread | GraphQL mutations `resolveReviewThread` / `unresolveReviewThread` — [GraphQL mutations](https://docs.github.com/en/graphql/reference/mutations#resolvereviewthread); resolvable by the PR author or anyone with write access — [Commenting on a PR](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/commenting-on-a-pull-request#resolving-conversations) | parity | same author-resolvability as GitLab → `challenge` stays acknowledgement-only (ADR-0017 §3) on both forges |
| C3 all-threads-resolved merge gate | branch protection "Require conversation resolution before merging" (also available in rulesets) — [About protected branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches#require-conversation-resolution-before-merging) | parity | gate absent/off → no deferred auto-merge (same doctor precondition as GitLab C3) |
| C4 approve / blocking review | `POST /repos/{o}/{r}/pulls/{n}/reviews` with `event: APPROVE` / `REQUEST_CHANGES` — [Reviews API](https://docs.github.com/en/rest/pulls/reviews?apiVersion=2022-11-28#create-a-review-for-a-pull-request) | partial | **REQUEST_CHANGES is a stronger primitive than any GitLab thread**: with required reviews it blocks merge until the same reviewer re-approves or the review is dismissed (see C8′). No `sha`-pin parameter on review creation — record `commit_id` from the response instead |
| C5 read who approved | `GET /repos/{o}/{r}/pulls/{n}/reviews` → `user.{id,login}`, `state`, `commit_id` — [Reviews API](https://docs.github.com/en/rest/pulls/reviews?apiVersion=2022-11-28#list-reviews-for-a-pull-request); aggregate: GraphQL `PullRequest.reviewDecision` (`APPROVED` / `CHANGES_REQUESTED` / `REVIEW_REQUIRED`) — [GraphQL Pulls](https://docs.github.com/en/graphql/reference/pulls#pullrequestreviewdecision) | parity | `commit_id` on each review is the per-approval SHA pin GitLab lacks (GitLab has only `approved_at` + reset-on-push) |
| C6 required approvals enforced | branch protection "Require pull request reviews before merging" (approving review count, optionally from code owners) — [About protected branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches#require-pull-request-reviews-before-merging) | parity (public repos) / partial (private repos need a paid plan) | absent → same consequence as GitLab Free: `require-review` never satisfiable → no auto-merge for archetypes needing it |
| C7 eligibility evidence chain | see §2 | partial | GitHub exposes the *aggregate* decision but not per-rule eligible-principal lists; CODEOWNERS matching is client-side |
| C8 self-approval limits | authors cannot review their own PR (the review UI/API is for reviewers; required reviews count "people with write permissions" other than the author — [About protected branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches#require-pull-request-reviews-before-merging)); "Require approval of the most recent reviewable push" excludes the pusher; "Dismiss stale approvals" resets on diff-changing pushes | partial | the docs state author-exclusion indirectly; the adapter must still exclude author/bot identities client-side from evidence (same rule as GitLab). **Uncited edge**: exact API status code when an author attempts self-approval — verify at E10 |
| C8′ who can dismiss the bot's REQUEST_CHANGES (OQ-18 core) | `PUT /repos/{o}/{r}/pulls/{n}/reviews/{id}/dismissals` (`event: DISMISS`) — [Reviews API](https://docs.github.com/en/rest/pulls/reviews?apiVersion=2022-11-28#dismiss-a-review-for-a-pull-request): "To dismiss a pull request review on a protected branch, you must be a repository administrator or be included in the list of people or teams who can dismiss pull request reviews"; unprotected default: "anyone with write permissions … can dismiss the blocking review" — [About protected branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches#require-pull-request-reviews-before-merging) | partial | dismissal is audit-trailed (review state `DISMISSED`) but, unrestricted, lets any write-access user (incl. the author on many setups) unblock without addressing findings. **Port requirement**: doctor on GitHub must verify dismissal restrictions exclude PR authors, else the challenge gate is author-bypassable — weaker than GitLab, where resolving threads at least leaves them individually audit-trailed and re-blockable |
| C9 bot identity | GitHub App installation token; review/comment author is the app's bot user (`<app>[bot]`), distinguishable by `user.id`/`user.type: Bot`; fine-grained least-privilege permissions per repo | parity | GITHUB_TOKEN inside Actions is also app-shaped; identity via `GET /user` equivalent (`GET /app` for apps). Exclusion logic identical to GitLab: match `user.id` of the token identity |
| C10 SHA-guarded merge (source CAS) | `PUT /repos/{o}/{r}/pulls/{n}/merge` with `sha` — "SHA that pull request head must match to allow merge"; `409 - Conflict if sha was provided and pull request head did not match`; `405` if not mergeable — [Pulls API](https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28#merge-a-pull-request) | parity | exact analogue of GitLab `merge?sha=` incl. the 405/409 split |
| C11 auto-merge arming | GraphQL `enablePullRequestAutoMerge` (input `pullRequestId`, `expectedHeadOid`, `mergeMethod`) / `disablePullRequestAutoMerge` — [GraphQL mutations](https://docs.github.com/en/graphql/reference/mutations#enablepullrequestautomerge); repo must "Allow auto-merge" — [Managing auto-merge](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/managing-auto-merge-for-pull-requests-in-your-repository); auto-merge disabled "if someone without write permissions pushes new changes" — [Automatically merging a PR](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/automatically-merging-a-pull-request) | partial | `expectedHeadOid` gives head pinning at arm time — but a **write-access** push does *not* auto-disarm (unlike GitLab, where any new commit cancels); protection is instead "dismiss stale approvals" + required checks re-running on the new head. Port must model "revocation trigger" per forge, not assume GitLab's |
| C12 merge-readiness polling | REST `mergeable_state` (undocumented enum) / GraphQL `reviewDecision` + `mergeStateStatus` | partial | fewer typed states than `detailed_merge_status`; adapter maps conservatively (unknown → not ready) |
| C13 evaluate merge result | `pull_request` events run CI on the PR **merge commit** (`GITHUB_SHA` = "last merge commit on the GITHUB_REF branch", ref `refs/pull/N/merge`) — [Events that trigger workflows](https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#pull_request) | parity | GitHub evaluates the merge result *by default* — stronger default than GitLab Free |
| C14 merge queue (merge-result pinning, OQ-18) | branch protection "Require merge queue"; queue builds temporary `gh-readonly-queue/{base}/…` branches containing target + queued PRs, requires checks via the `merge_group` event, drops entries whose group fails — [Managing a merge queue](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/managing-a-merge-queue) | parity | direct analogue of merge trains → satisfies ADR-0017 §1 merge-result pinning. Availability: org repos (public; private/internal on Enterprise Cloud) → outside that: capability gap → no deferred auto-merge |
| C15 queue enforcement | "Require merge queue" is itself a protection rule: once required, merges route through the queue | parity | admin bypass exists unless "Do not allow bypassing" is set — doctor-reportable, same as GitLab |
| C16 read merge-result commit | `refs/pull/{n}/merge` ref / `merge_commit_sha` on the PR object | partial | usable as recorded digest; semantics of `merge_commit_sha` are state-dependent — verify at E10 |
| C17 protected pipeline source (fork/base-ref trust, ADR-0015 §4) | `pull_request` from forks: workflow runs with **read-only `GITHUB_TOKEN`, no secrets** — [Events: forked repositories](https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#workflows-in-forked-repositories); `pull_request_target` "runs in the context of the default branch of the base repository" with privileges and **must not check out untrusted code** — [same doc](https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#pull_request_target) + [Actions security hardening](https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions) | partial | matches ADR-0015 §8 row "CI on fork → advisory-only": on `pull_request` the definition is author-influenced but unprivileged (advisory-only is safe); privileged gating needs `pull_request_target`/`workflow_run` **without** checking out untrusted code, or same-repo branches + branch protection on workflow paths |
| C18 external status checks | required status checks (commit statuses / check runs), optionally pinned to one GitHub App as the expected source — [About protected branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches#require-status-checks-before-merging) | parity | app-pinning is *stronger* than GitLab's Ultimate-only status checks (source authenticity) |
| C19 approval-reset on push | branch protection "Dismiss stale pull request approvals when new commits are pushed" (records diff state; also dismisses when the merge base introduces new changes) — [About protected branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches#require-pull-request-reviews-before-merging) | parity | opt-in (GitLab's is default-on) → doctor must verify it, plus "Require approval of the most recent reviewable push" for the strongest variant |

## 2. `require-review` evidence chain on GitHub (OQ-23 counterpart)

| Step | Call | Caveat |
| --- | --- | --- |
| (a) required rules | `GET /repos/{o}/{r}/branches/{branch}/protection` (`required_pull_request_reviews.{required_approving_review_count, require_code_owner_reviews, dismissal_restrictions}`) or rulesets `GET /repos/{o}/{r}/rules/branches/{branch}` | needs admin-ish read scope; rulesets and legacy protection must both be checked |
| (b) eligible principals | CODEOWNERS file **from the base ref** (fork-safe by definition — [About code owners: CODEOWNERS and forks](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners#codeowners-and-forks)) + team expansion `GET /orgs/{org}/teams/{team_slug}/members`; owners must hold write — [About code owners](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners) | **no API returns the computed per-PR eligible code owners** — pattern matching (gitignore-like, last-match-wins, any-one-owner-suffices) is reimplemented client-side → typed-eligible-principals evidence is `partial` |
| (c) actual approvals | `GET /pulls/{n}/reviews` (`user`, `state`, `commit_id`) + GraphQL `reviewDecision` as the forge's own verdict | `reviewDecision: APPROVED` proves the forge considers requirements met without naming *which* rule each review satisfied — record both the aggregate and the identity list |

Adversarial cases: author self-approval (rejected by the platform for required reviews; exclude
by `user.id == PR author id` regardless) and assent's own bot approval (exclude by token
identity `user.id`; an app cannot satisfy "required approving reviews" as a human principal —
verify exact behaviour at E10). Verdict for ADR-0017 §3: GitHub can prove *that* eligible
approval exists (aggregate `reviewDecision` + protection config) — `partial` on proving
*which* typed principal satisfied *which* rule.

## 3. Arm-and-wait flow reproduction (OQ-7 / OQ-18 answer)

GitLab flow (ADR-0009 amendment): resolvable threads + all-resolved gate + approve +
`auto_merge` pinned via `merge?sha=`; forge merges when threads resolve; new push cancels.

GitHub reproduction: findings as review threads under one **`REQUEST_CHANGES` review** +
branch protection with required reviews, **required conversation resolution**, required
status checks (+ merge queue where available) + `enablePullRequestAutoMerge(expectedHeadOid)`.
Merge happens only when: conversations resolved, the bot's blocking review is lifted, checks
pass. **On paper: yes — parity, with three port-relevant deltas:**

1. **Unblocking is review-dismissal or bot re-review, not thread resolution.** Resolving all
   threads does not clear `CHANGES_REQUESTED`; someone must dismiss the bot's review (C8′) or
   the bot must submit a superseding APPROVE. Cleanest mapping: assent re-arms by submitting
   APPROVE itself once its *own* conditions are met — but one-shot CI cannot observe
   resolution (same limitation as GitLab, ADR-0009 amendment), so on GitHub the
   acknowledgement gate rests on **required conversation resolution (C3) alone**, and the
   REQUEST_CHANGES device is reserved for `block`-severity findings (which never auto-merge
   anyway). This asymmetry is the honest answer to OQ-7: parity for the *gate*, not for the
   *device*.
2. **Revocation semantics differ (C11)**: GitLab cancels the armed merge on any new commit;
   GitHub keeps auto-merge armed on write-access pushes and relies on stale-approval dismissal
   + re-running checks. The port's `ArmDeferredMerge` contract must therefore state the
   revocation trigger it guarantees, and the GitHub adapter must require
   dismiss-stale-approvals + (where available) merge queue to reach GitLab-equivalent safety.
3. **Merge queue = merge-result pinning (C14)**: with a queue, target movement re-validates
   the merge result (ADR-0017 §1 satisfied); without one, GitHub has "Require branches to be
   up to date" (strict checks) as a weaker substitute (forces re-run on stale head, but the
   author chooses when to update) — deferred auto-merge without queue on a busy target =
   capability gap, fail closed.

## 4. Port-design consequences (keeping the seam honest)

- Capability flags the Forge port needs (superset check against ADR-0005): `resolvable-threads`,
  `threads-block-merge`, `blocking-review`, `review-dismissal-restrictions`,
  `sha-guarded-merge`, `deferred-merge-arming`, `arming-revoked-on-push`,
  `merge-result-pinning`, `eligible-approval-evidence{full|aggregate}`,
  `approval-reset-on-push{default|opt-in}`, `protected-pipeline-source`.
- Two GitHub-only behaviours the port must not preclude: review lifecycle (submit → dismiss →
  re-request) and GraphQL-only thread resolution (adapter needs a GraphQL client).
- Nothing in the GitLab dossier's §2 precondition table is GitLab-only: source-CAS, queue
  pinning, and record-only merge-ref all have GitHub analogues → the Reconcile precondition
  shape freezes forge-neutrally.

## Open verification items (E10-era, not scheduled)

- Author self-approval and bot-as-required-reviewer exact API responses (§2 adversarial).
- `merge_commit_sha`/`refs/pull/N/merge` freshness semantics (C16).
- Plan gating re-check: merge queue and protection features per repo visibility/plan.
- Whether `enablePullRequestAutoMerge` requires the PR to be "not currently mergeable"
  (UI shows the option only then — API behaviour when already mergeable needs a live check).
