# Forge behaviour dossier — GitLab (P1-E3-S01/S02)

Verified API mechanics for the Forge/Publisher/Reconcile port design (ADR-0005, ADR-0009
amendment, ADR-0015 §2/§4/§5, ADR-0017 §1/§3/§7). Every claim carries an endpoint citation
(method + path + doc URL) and, per capability, the GitLab tier (Free/Premium/Ultimate) it is
available on plus the **fail-closed consequence of absence**. Docs snapshot: docs.gitlab.com,
2026-07-21. Companion: [forge-dossier-github.md](forge-dossier-github.md) (parity mapping).

Legend: F = Free, P = Premium, U = Ultimate. "Free" claims below refer to the licensed tier;
where CE (unlicensed EE/Community Edition) semantics differ, the row says so explicitly.

## 1. Capability matrix

| # | Capability | Mechanism (method + endpoint) | Tier | Fail-closed consequence of absence |
| --- | --- | --- | --- | --- |
| C1 | Post resolvable finding thread | `POST /projects/:id/merge_requests/:iid/discussions` (plain or diff-positioned via `position[*]`); notes carry `resolvable: true` — [Discussions API](https://docs.gitlab.com/api/discussions/#create-a-merge-request-thread) | F/P/U | n/a (present on all tiers) |
| C2 | Resolve / reopen a thread | `PUT /projects/:id/merge_requests/:iid/discussions/:discussion_id?resolved=` — resolvable by Developer+ **or the MR author** — [Discussions API](https://docs.gitlab.com/api/discussions/#resolve-a-merge-request-thread) | F/P/U | n/a. Note: author-resolvability is why `challenge` is acknowledgement only (ADR-0017 §3) |
| C3 | All-discussions-resolved merge gate | project setting `only_allow_merge_if_all_discussions_are_resolved` (`PUT /projects/:id`); surfaced as `detailed_merge_status: discussions_not_resolved` — [Projects API](https://docs.gitlab.com/api/projects/#edit-a-project), [MR merge status](https://docs.gitlab.com/api/merge_requests/#merge-status) | F/P/U | gate absent → resolution has no merge-blocking force → **no deferred auto-merge**; `assent doctor` must verify the setting before arming (ADR-0009 amendment) |
| C4 | Approve / unapprove MR | `POST /projects/:id/merge_requests/:iid/approve` (optional `sha=` — mismatch → `409 Conflict`); `POST .../unapprove` — [Approvals API](https://docs.gitlab.com/api/merge_request_approvals/#approve-merge-request) | F/P/U | n/a. On Free, approvals are **optional and never merge-blocking** (see C6) |
| C5 | Read who approved (identities) | `GET /projects/:id/merge_requests/:iid/approvals` → `approved_by[].user.{id,username}`, `approved_at`, `approvals_required/left`, `approved` — [Approvals API](https://docs.gitlab.com/api/merge_request_approvals/#retrieve-approval-state-for-a-merge-request) | F/P/U | n/a. CE caveat: `approved` is `true` when **at least one** approval exists (no rule semantics); EE: `true` when configured rules are satisfied |
| C6 | Required approval rules (enforced) | project/MR approval rules; `detailed_merge_status: not_approved` blocks merge — [Approval rules](https://docs.gitlab.com/user/project/merge_requests/approvals/rules/), [Approvals overview](https://docs.gitlab.com/user/project/merge_requests/approvals/) ("GitLab Free … approvals are optional and don't prevent merging") | P/U | **capability gap → `require-review` never satisfiable → no auto-merge for archetypes needing it** (OQ-23; §4 below) |
| C7 | Approval rules / eligibility APIs | `GET .../approval_rules`, `GET .../approval_state` (per-rule `eligible_approvers[]`, `approved_by[]`), `GET/POST /projects/:id/approvals` — [Approvals API](https://docs.gitlab.com/api/merge_request_approvals/#approval-rules-for-a-merge-request) ("All other endpoints require Premium or Ultimate") | P/U | same as C6 — without the evidence chain the adapter cannot *prove* eligibility, which is the ADR-0017 §3 requirement |
| C8 | Self-approval limits | `POST /projects/:id/approvals` fields `merge_requests_author_approval`, `merge_requests_disable_committers_approval`, `require_reauthentication_to_approve`; UI defaults documented as "By default, the creator of a merge request (author) cannot approve it" — [Approval settings](https://docs.gitlab.com/user/project/merge_requests/approvals/settings/), [Approvals API](https://docs.gitlab.com/api/merge_request_approvals/#update-approval-configuration-for-a-project) | P/U (settings page and config API are Premium) | on Free the forge does not enforce author/committer exclusion → the adapter must exclude author/bot approvals **client-side** from evidence; since C6 is absent anyway, `require-review` is already unsatisfiable on Free |
| C9 | Bot identity (write actor) | project access token → GitLab creates a **bot user** member, username `project_{project_id}_bot_{random_string}` — [Project access tokens](https://docs.gitlab.com/user/project/settings/project_access_tokens/#bot-users-for-projects); resolve own identity via `GET /user` — [Users API](https://docs.gitlab.com/api/users/#retrieve-the-authenticated-user) | GitLab.com: **P/U only** ("On GitLab.com, project access tokens require a Premium or Ultimate subscription"). Self-Managed/Dedicated: any license | on GitLab.com Free, no project token → bot identity must be a dedicated user account with PAT (heavier custody) or **no write actions at all** (advisory-only mode, ADR-0015 §8) |
| C10 | SHA-guarded merge (source CAS) | `PUT /projects/:id/merge_requests/:iid/merge?sha=` — "this SHA must match the HEAD of the source branch" — [MR API](https://docs.gitlab.com/api/merge_requests/#merge-a-merge-request); response codes in §3 | F/P/U | adapter that cannot SHA-guard must declare the gap → engine never auto-merges (ADR-0015 §2). Present on all tiers, so no gap on GitLab |
| C11 | Auto-merge arming ("merge when checks pass") | `PUT .../merge` with `auto_merge=true` (17.11+; `merge_when_pipeline_succeeds` deprecated) **combined with** `sha=`; cancel: `POST .../cancel_merge_when_pipeline_succeeds` (also removes from merge train) — [MR API](https://docs.gitlab.com/api/merge_requests/#merge-a-merge-request), [Cancel](https://docs.gitlab.com/api/merge_requests/#cancel-merge-when-pipeline-succeeds); behaviour: new commits **cancel the armed merge** — [Auto-merge](https://docs.gitlab.com/user/project/merge_requests/auto_merge/) | F/P/U | n/a — this is the ADR-0009-amendment arming primitive; preconditions (C3 gate on, pipelines-must-succeed) verified by `doctor` |
| C12 | Merge-readiness polling | `GET .../merge_requests/:iid` → `detailed_merge_status` enum incl. `discussions_not_resolved`, `not_approved`, `ci_must_pass`, `requested_changes`, `need_rebase`, `approvals_syncing` — [MR API merge status](https://docs.gitlab.com/api/merge_requests/#merge-status) | F/P/U (some statuses only fire with P/U features) | n/a |
| C13 | Merged results pipelines (evaluate the merge result) | project setting; pipeline runs on "a temporary merged commit that combines code from the source and target branches" — [Merged results pipelines](https://docs.gitlab.com/ci/pipelines/merged_results_pipelines/) | P/U | **capability gap → evaluation sees the source branch only, never the merge result → merge-result digest precondition (ADR-0017 §7) unverifiable → no deferred auto-merge on this tier** |
| C14 | Merge trains / queue (merge-result pinning) | settings + `POST /projects/:id/merge_trains/merge_requests/:iid` (supports `sha=`, `auto_merge=`; `201/202`); status `GET /projects/:id/merge_trains/...` — [Merge trains](https://docs.gitlab.com/ci/pipelines/merge_trains/), [Merge trains API](https://docs.gitlab.com/api/merge_trains/) | P/U | **capability gap → no forge mechanism re-validates against a moved target → per ADR-0017 §1, deferred auto-merge is refused on this tier** (immediate merge with fresh re-evaluation remains possible) |
| C15 | Merge train enforcement (no bypass) | Settings > Merge requests > "Merge train enforcement" (Allow bypass / Enforce for all users / Enforce with Owner override); enforced mode "rejects direct merges through the REST API and GraphQL API" — [Merge trains: enforce](https://docs.gitlab.com/ci/pipelines/merge_trains/#enforce-merge-trains) | P/U | without enforcement, humans can `Merge immediately` around the train — acceptable (human action ≠ assent's write path), but doctor should report it |
| C16 | Read merge-result commit (evidence, not enforcement) | `GET .../merge_requests/:iid/merge_ref` → `{"commit_id": …}` — updates `refs/merge-requests/:iid/merge` to "the state the target branch would have if a regular merge action was taken" — [MR API](https://docs.gitlab.com/api/merge_requests/#merge-to-default-merge-ref-path) | F/P/U | n/a — usable on every tier to *record* the evaluated merge-result digest in `Decision.Pins`; enforcement still needs C13/C14 |
| C17 | Protected pipeline source (job def not author-editable) | (a) CI/CD config file in **another project** (`path/file.yml@group/project`, Settings > CI/CD) — [Pipeline settings](https://docs.gitlab.com/ci/pipelines/settings/#specify-a-custom-cicd-configuration-file) (F/P/U); (b) **pipeline execution policies** (inject/override strategies, `.pipeline-policy-pre` reserved stage, `skip_ci` control) — [Pipeline execution policies](https://docs.gitlab.com/user/application_security/policies/pipeline_execution_policies/) (U); compliance pipelines are **deprecated** in favour of (b) (same doc) | (a) F/P/U · (b) U | if neither is used and the assent job lives in the MR-editable `.gitlab-ci.yml`, the topology is **unsupported/insecure** (ADR-0015 §4); `doctor` refuses to arm auto-merge (ADR-0015 §8) |
| C18 | External status checks as merge gate | project setting `only_allow_merge_if_all_status_checks_passed` — [Projects API](https://docs.gitlab.com/api/projects/) ("Ultimate only") | U | optional defense-in-depth only; absence is not a gap for the v1 design |
| C19 | Bot approval-reset immunity controls | `reset_approvals_on_push` (default resets approvals on new push, patch-id based), `selective_code_owner_removals`; bot-only `PUT .../reset_approvals` ("Available only to bot users with a valid project or group token") — [Approvals API](https://docs.gitlab.com/api/merge_request_approvals/#reset-approvals-for-a-merge-request) | reset-on-push behaviour F/P/U; config API P/U | approvals resetting on push is the **desired** fail-closed direction (stale approval cannot survive new commits); `doctor` should verify `reset_approvals_on_push: true` |

## 2. Reconcile preconditions (ADR-0017 §7) → GitLab mechanism per tier

Preconditions carried by `Reconcile(DesiredReviewState, Preconditions)`: **source SHA, target
SHA, evaluated merge-result digest**, fact validity deadline, decision hash.

| Precondition | Enforce/verify mechanism | Free | Premium/Ultimate |
| --- | --- | --- | --- |
| Source SHA unchanged | `PUT .../merge?sha=` CAS (C10); `POST .../approve?sha=` (C4); armed auto-merge cancelled on new push (C11) | enforced | enforced |
| Target SHA unchanged / re-validated | **no target-SHA CAS parameter exists on the merge endpoint** (verified against the [MR API attribute table](https://docs.gitlab.com/api/merge_requests/#merge-a-merge-request) — `sha` refers to the *source* HEAD only). Only a merge train re-validates after target movement (C14) | **capability gap → no deferred auto-merge on this tier** | merge train (C14) re-runs the merged-result pipeline when the train/target changes |
| Evaluated merge-result digest | record: `GET .../merge_ref` (C16, all tiers); enforce: merged-results pipeline + merge train (C13+C14) — the train pipeline runs on the merge result and drops the MR if it becomes invalid | record-only → **capability gap → no deferred auto-merge on this tier** | enforced via train |
| Fact validity deadline | no forge mechanism — assent-side arming precondition (`facts.max_age`, ADR-0017 §4): one-shot may arm only if obligations cannot expire before a forge-enforced event | assent-side | assent-side |

Fail-closed summary (ADR-0017 §1 wording): `merge?sha=` alone is **source-only CAS and
insufficient**; on tiers without merge trains the engine either merges immediately in the same
run that evaluated (target pinned by comparing `diff_refs.base_sha`/`merge_ref` before the
write) or refuses deferred auto-merge: `capability gap → no deferred auto-merge on this tier`.

## 3. SHA-guard behaviour: `PUT /merge_requests/:iid/merge?sha=` (REQ-P1-E3-S01-02)

Documented contract ([MR API](https://docs.gitlab.com/api/merge_requests/#merge-a-merge-request)):
`sha` — "If present, this SHA must match the HEAD of the source branch. Use to ensure that
only reviewed commits are merged." Documented failure codes:

| HTTP | Message | Meaning for the adapter |
| --- | --- | --- |
| `401` | `401 Unauthorized` | token lacks merge permission → fail closed, report capability/permission error |
| `405` | `405 Method Not Allowed` | MR "cannot merge" (draft, blocked, gate unmet) → fail closed, re-read `detailed_merge_status` |
| `409` | `SHA does not match HEAD of source branch` | **the CAS miss** — HEAD moved after evaluation → decision void, re-evaluate |
| `422` | `Branch cannot be merged` | merge itself failed (e.g. conflict) → fail closed |

Adversarial case (ADR-0015 §2): evaluate at `SHA_A` → author pushes `SHA_B` → adapter calls
`merge?sha=SHA_A` → expected `409` + no merge; the new push has also cancelled any armed
auto-merge (C11) and (default) reset approvals (C19). The same guard exists on approval:
`approve?sha=` mismatch → `409 Conflict` ([Approvals API](https://docs.gitlab.com/api/merge_request_approvals/#approve-merge-request)).

| Verification | Status |
| --- | --- |
| Docs-derived expectation (table above) | done (this section) |
| Live gitlab.com run: push-after-evaluation → pinned merge attempt → capture actual codes/bodies | **live verification pending — supplied by P2-E2 smoke (spike B)** (no working gitlab.com token available in this lane; decision logged in the workspace INBOX) |

Edge to capture in the live run: response code when `sha=` matches but the **target** moved
(expected: merge succeeds — which is exactly the ADR-0017 §1 insufficiency being proven), and
`merge_when_pipeline_succeeds`-deprecation behaviour of `auto_merge=true` on 18.x.

## 4. Approval-eligibility evidence for `require-review` (S02, OQ-23)

ADR-0017 §3: `require-review` is satisfied only by **forge-proven eligible approval**; failed
authorization never degrades into an author-resolvable thread.

### Premium / Ultimate — the evidence chain

| Step | Call | Fields consumed |
| --- | --- | --- |
| (a) required rules | `GET /projects/:id/merge_requests/:iid/approval_rules` ([doc](https://docs.gitlab.com/api/merge_request_approvals/#list-all-approval-rules-for-a-merge-request), P/U) | `id`, `name`, `rule_type` (`regular`, `code_owner`, `report_approver`, `any_approver`), `approvals_required`, `overridden` |
| (b) eligible approvers | same response / `GET .../approval_state` ([doc](https://docs.gitlab.com/api/merge_request_approvals/#retrieve-approval-details-for-a-merge-request), P/U) | `rules[].eligible_approvers[].{id,username}` — CODEOWNERS-sourced rules appear as `rule_type: code_owner` (typed eligible principals, ADR-0017 §3) |
| (c) actual approvals + identities | `GET .../approval_state` per rule; `GET .../approvals` ([doc](https://docs.gitlab.com/api/merge_request_approvals/#retrieve-approval-state-for-a-merge-request), F) for the flat list | `rules[].approved_by[].{id,username}`, `rules[].approved`; `approved_by[].user.{id,username}` + `approved_at` |

Adversarial exclusions — exact fields:

- **MR-author approval**: compare each `approved_by[].user.id` against the MR `author.id`
  (`GET /projects/:id/merge_requests/:iid`, [doc](https://docs.gitlab.com/api/merge_requests/#get-single-mr)).
  Forge-side enforcement: `merge_requests_author_approval: false` in
  `GET /projects/:id/approvals` ([doc](https://docs.gitlab.com/api/merge_request_approvals/#retrieve-approval-configuration-for-a-project));
  the adapter must verify the setting **and** exclude client-side (defense in depth, since the
  setting is overridable per-MR unless `disable_overriding_approvers_per_merge_request`).
- **assent's own bot approval**: the token's identity from `GET /user`; project-access-token
  actors are bot users with username `project_{project_id}_bot_{random_string}`
  ([doc](https://docs.gitlab.com/user/project/settings/project_access_tokens/#bot-users-for-projects)) —
  distinguishable by exact `user.id` match (primary) and the `project_*_bot_*` username shape
  (sanity check). Excluded from evidence for *any* rule: assent approving cannot satisfy
  `require-review` (ADR-0015 §5 makes the residual-gate implication explicit).
- **Committer approval**: `merge_requests_disable_committers_approval` (same config API) if
  the archetype demands author-independence stronger than authorship alone.
- **Invalid-rule pitfall (must-check)**: GitLab marks rules "Auto approved" when they are
  *impossible to satisfy* (only eligible approver is the author; no eligible approvers;
  required > eligible) — [Invalid rules](https://docs.gitlab.com/user/project/merge_requests/approvals/#invalid-rules).
  A rule auto-approved this way has `approved: true` with empty/insufficient `approved_by` —
  the adapter must require **non-empty eligible `approved_by`**, never trust the rule-level
  boolean alone.
- **Staleness**: approvals reset on new push by default (`reset_approvals_on_push`, patch-id
  based — [settings](https://docs.gitlab.com/user/project/merge_requests/approvals/settings/#remove-all-approvals-when-commits-are-added-to-the-source-branch));
  pair `approved_at` with the evaluated head SHA, and treat `detailed_merge_status:
  approvals_syncing`/`checking` as not-yet-provable ([timing note](https://docs.gitlab.com/api/merge_request_approvals/#prevent-approval-resets-in-automated-merge-requests)).
- **Re-authentication signal**: `require_reauthentication_to_approve` strengthens the identity
  claim (CFR-11-style e-signature) — recordable in ApprovalEvidence but not required for v1.

### Free — capability gap

`GET .../approvals` exists (identities visible), but: approval **rules** APIs are P/U-only
([Approvals API tier note](https://docs.gitlab.com/api/merge_request_approvals/)), required
approvals are P/U ([Approvals overview](https://docs.gitlab.com/user/project/merge_requests/approvals/#required-approvals)),
and on Free "approvals are optional and don't prevent merging" (same doc). There is no
forge-proven notion of *eligible* approver and no merge-blocking force. Additionally, CE's
`approved` field is `true` on **any** single approval ([doc](https://docs.gitlab.com/api/merge_request_approvals/#retrieve-approval-state-for-a-merge-request)).

Recorded consequence: **capability gap → `require-review` never satisfiable → no auto-merge
for archetypes needing it** — never a silent downgrade to a resolvable thread (ADR-0017 §3).

## 5. Merge trains as merge-result pinning (ADR-0017 §1)

- Train pipelines run on "the changes of the merge requests combined with the target branch";
  a failed/invalidated entry is dropped and later pipelines restart against the new combined
  state — [Merge trains](https://docs.gitlab.com/ci/pipelines/merge_trains/) (P/U).
- API arming: `POST /projects/:id/merge_trains/merge_requests/:iid` with `sha=` (source CAS at
  train entry) + `auto_merge=true` → `201/202` — [Merge trains API](https://docs.gitlab.com/api/merge_trains/#add-a-merge-request-to-a-merge-train).
- Drop conditions relevant to assent: draft, merge conflict, and "a new conversation thread
  that is unresolved, when all threads must be resolved is enabled" — i.e. the C3 gate keeps
  protecting the armed state *inside* the train ([troubleshooting](https://docs.gitlab.com/ci/pipelines/merge_trains/#merge-request-dropped-from-the-merge-train)).
- Bypass control: merge-train enforcement modes (C15). `skip_merge_train=true` on the merge
  endpoint is the documented bypass — assent must never use it.
- Constraint: with trains enabled, plain auto-merge cannot skip the train ([cannot use auto-merge](https://docs.gitlab.com/ci/pipelines/merge_trains/#cannot-use-auto-merge)) —
  the adapter's arming call *is* the train-add on these projects.

## 6. Protected-pipeline topologies (ADR-0015 §4)

| Topology | Tier | Mechanics | doctor verdict |
| --- | --- | --- | --- |
| CI config file outside the repo (`.yml@group/project` or external URL) | F/P/U | Settings > CI/CD > "CI/CD configuration file"; config project can be locked down ("Give write permissions on the project only to users who are allowed to edit the file") — [doc](https://docs.gitlab.com/ci/pipelines/settings/#specify-a-custom-cicd-configuration-file) | supported — **the one v1 path available on every tier** (feeds OQ-24) |
| Pipeline execution policies (security policy project) | U | `pipeline_execution_policy` with `inject_policy`/`override_project_ci`, reserved `.pipeline-policy-pre` stage, `skip_ci` ignored by default — [doc](https://docs.gitlab.com/user/application_security/policies/pipeline_execution_policies/) | supported (strongest: MR authors cannot remove the job) |
| Compliance pipelines | U (deprecated) | "Compliance pipelines are deprecated… use pipeline execution policies for all new implementations" — [doc](https://docs.gitlab.com/user/application_security/policies/pipeline_execution_policies/) | migrate; do not document as a target topology |
| assent job in the MR-editable `.gitlab-ci.yml` with a privileged token | any | author can edit/remove the gate and exfiltrate the token | **unsupported, insecure** — doctor refuses to arm auto-merge (ADR-0015 §8) |

Residual caveat (all tiers): even with an external config file, `workflow:rules`/`rules` in
included author-editable files can suppress jobs; the policy-strategy (`override_project_ci`)
is the only variant immune to that. The E4 adapter spec must pin the exact include layout.

## 7. Auto-merge arming lifecycle (ADR-0009 amendment)

1. **Preconditions checked by `doctor`**: C3 gate on; "Pipelines must succeed" on
   ([auto-merge prerequisites](https://docs.gitlab.com/user/project/merge_requests/auto_merge/));
   protected topology (§6); P/U: trains enabled for merge-result pinning (§5).
2. **Arm**: approve (`sha=`-pinned) + `PUT .../merge?auto_merge=true&sha=<evaluated>` (or the
   train-add on train projects). GitLab then merges only when **all merge checks** pass —
   including "All discussions are resolved" and "All required approvals" ([auto-merge checks list](https://docs.gitlab.com/user/project/merge_requests/auto_merge/)).
3. **Forge-native revocation** (what makes one-shot arming sound, ADR-0017 §4): new source
   commits cancel the armed merge and (default) reset approvals; on ff-only projects target
   movement also cancels ([pipeline success behaviour](https://docs.gitlab.com/user/project/merge_requests/auto_merge/#pipeline-success-for-auto-merge)); on trains, invalidation drops the entry (§5).
4. **What the forge does NOT revoke**: fact expiry (`facts.max_age`) — assent-side arming
   precondition; decisions with expiring authorization facts must not arm (ADR-0017 §4).
5. **Cancel path** (assent- or human-initiated): `POST .../cancel_merge_when_pipeline_succeeds`
   → `201`; `406` when the MR is closed ([doc](https://docs.gitlab.com/api/merge_requests/#cancel-merge-when-pipeline-succeeds)).

## Open verification items

- Live `merge?sha=` run on gitlab.com (§3) — **pending, P2-E2 smoke (spike B)**.
- Free-tier author-self-approval default (does CE block author approval without the Premium
  settings page?) — irrelevant to gating (C6 gap already refuses `require-review` on Free) but
  worth confirming in the same smoke run.
- `auto_merge=true` + `sha=` interaction on merge trains (does the train re-check the source
  CAS at final merge, or only at train entry?) — E4 conformance case.
