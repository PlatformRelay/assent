# ADR-0015: Trust boundaries and merge-time integrity

| | |
| --- | --- |
| **Status** | Proposed |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0007 effects](0007-rule-effects-decision-aggregation.md) · [ADR-0008 routing](0008-change-classification-routing-scope.md) · [ADR-0009 modes](0009-execution-modes.md) · [ADR-0010 config](0010-config-files-repo-layout.md) · [ADR-0011 ports](0011-core-ports-and-contracts.md) · adversarial design review 2026-07-21 (findings F1, F3, F4, F5, F13, F14) |

## Context

An adversarial review of the design found that the trust model was implicit — and, as
written, broken: policies were loaded from the very branch they gate (F1), the merge action
was not pinned to the evaluated commit (F3), facts could go stale between evaluation and a
deferred merge (F4), the CI job definition itself is author-editable in self-service repos
(F5), and the webhook mode deferred its security requirements while claiming
architecture-readiness (F13). This ADR fixes the trust boundary explicitly. Principle:
**everything that decides is loaded from a trusted ref; everything that is judged comes from
the untrusted branch; every write action re-verifies what it is about to act on.**

## Decision (proposed)

### 1. Policy is trusted input → loaded from the target ref, never the MR branch

`.assent/**` (config, bindings, packs, referenced Rego, templates, tests' expectations) is
**always loaded from the target/base ref** of the MR. The source branch contributes only the
material under judgment: the diff and the branch state for `scope: branch` rules.

Any MR that touches `.assent/**` routes to the implicit meta-class **`assent-policy`**,
which is `block`-by-default: policy changes always require human review and can never be
vouched by the policies they modify. Repos may relax this only to `challenge`, never to
`vouch`. A mandatory golden e2e case: *"MR edits its own policy → BLOCK."*
(Recommended hardening: forge-level CODEOWNERS/approval rule on `.assent/**` requiring a
human — see §5.)

### 2. Merge is SHA-guarded (no TOCTOU)

`Publisher.Merge` and `Publisher.Approve` carry the evaluated head SHA from `Decision.Pins`
and use the forge's compare-and-swap: GitLab `merge?sha=`, GitHub merge `sha` parameter.
If HEAD has moved, the merge **fails closed** and the new pipeline re-evaluates. A forge
adapter that cannot SHA-guard merges must declare the capability gap, and the engine then
never auto-merges on that forge. This is a required case in the conformance suite (ADR-0005).

### 3. Fact freshness is bounded

`Decision.Pins` records per-provider resolution timestamps and the full resolved fact set
(which also makes replay hermetic — see ADR-0011 amendment). When the merge is deferred
(forge-native gate, ADR-0009), staleness is bounded by: (a) any new push re-runs the
pipeline and re-resolves facts (forge behavior), and (b) a configurable
`facts.max_age` (default: 24h) after which an armed auto-merge is considered expired — the
summary comment states the expiry and a pipeline re-run is required. Repos with
security-sensitive facts should set a short `max_age` or use serve mode (v1.x) which
re-evaluates on events.

### 4. The pipeline running assent must not be author-editable

CI-first only works if the job definition is trusted. Adoption prerequisite (documented in
the walkthrough and CI templates, verified by `assent doctor`): the assent job must come from
a **protected source** — GitLab: pipeline configuration from a protected/included file
outside the MR branch's control (compliance pipeline / instance template) or a
merge-request-approved pipeline model; GitHub: workflows from the target branch
(`pull_request` runs the base-ref workflow for forks; same-repo branches need branch
protection on workflow paths). The token is least-privilege: merge+comment on this project
only. The docs must state plainly: putting the assent job in an author-editable
`.gitlab-ci.yml` with a privileged token is an unsupported, insecure topology.

### 5. Bot approval is the gate — say it out loud

When assent approves and merges, its approval satisfies the forge's "1 approval required"
rule; there is **no residual human gate**. This is intended and must be documented, together
with the defense-in-depth option: a forge approval rule on `.assent/**` (and optionally on
`prod/**` classes) requiring an identity assent cannot provide.

### 6. Serve mode: security requirements fixed now, implementation later

So that v1 port shapes don't foreclose them (F13): webhook signature verification (HMAC /
forge-native), event dedup key (MR id + head SHA + event id), idempotent publishing (findings
carry stable ids; re-posting is an upsert), per-repo token scoping (no instance-wide
credential), and the same target-ref policy loading as §1.

## Consequences

- The classifier (ADR-0008) gains the built-in `assent-policy` meta-class; ADR-0010's layout
  section references §1; ADR-0014's expectation-trust question is resolved the same way:
  expectations for `assent test` in CI come from the target ref when `.assent/**` changed.
- `scan` backtests policies from the ref being tested, not from each historical MR — already
  consistent with §1.
- Publisher port signatures change: `Approve`/`Merge` take the pinned SHA (ADR-0011 amended).
- Threat model becomes documentable: this ADR seeds SECURITY.md.

## Counterpoints considered

- *"Loading policy from the target ref makes policy changes hard to test in their own MR."* —
  No: `assent test` and `--dry-run` run locally against the branch; CI additionally runs the
  branch's own tests *as tests* (L1). Only the *gating* decision uses target-ref policy. The
  policy MR is gated by the old policy + human review — exactly right.
- *"SHA-guard + freshness makes automerge slower."* — It makes automerge *checkable*. The
  fast path (APPROVE at evaluation time, merge immediately, SHA still HEAD) is unaffected.

## Additions (2026-07-21, independent security review A-03/A-04/A-05/A-10)

### 7. Provider/plugin trust model (A-03)

Providers run **without the forge write token** — the provider host passes a narrow FactQuery
and receives facts; the approve/merge credential never enters a provider's environment or
process. Exec/gRPC provider binaries are **digest-pinned** in `config.yaml` (optional cosign
verification); default recommendation is *no in-process third-party code* — subprocess with
narrow JSON RPC (tier 2) or WASM (tier 4) are the sanctioned paths, and the gRPC tier is
documented as elevated-risk until sandboxing lands. Provider timeouts fail closed (facts
`unknown` → fail-safe per ADR-0007 tri-state).

### 8. Execution-authority matrix (A-04)

| Mode / context | May hold write token | May auto-approve/merge |
| --- | --- | --- |
| CI from protected/trusted config (§4) | yes (least-privilege) | yes |
| CI on fork / untrusted-contributor MR | **no** | advisory-only (report, no writes) |
| `serve` (v1.x, §6) | yes (per-repo scoped) | yes |
| `--dry-run` / `explain` / `test` / `scan` | no | never |

`assent doctor` refuses to arm auto-merge when it cannot verify the protected-config
precondition.

### 9. Input resource safety & report retention (A-05, review P1-3)

Parsing and evaluation run under hard limits — max file size/count, max diff bytes, nesting
depth, YAML anchor/alias expansion caps, symlink and path-traversal rejection, parse+eval
deadlines, CEL cost budget — breach fails closed to REVIEW (spec'd with ADR-0003). The
summary comment always embeds the decision hash and a link to the report artifact; docs must
warn that CI artifact retention limits the audit window and recommend a retention policy for
report artifacts.
