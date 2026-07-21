# Later phases — epics only (stories authored in Phase 3, per meta-plan "contracts first")

Statuses: **Planned** (unlocks with its phase) · **Locked (D-012)** (needs a named consumer).

## Phase 3 — Contracts first

### P3-E1 — JSON Schemas + the strict contract fixture — Planned
Freeze the serialized contracts as versioned JSON Schemas — authored resources (Config,
RulesetBinding, MergePolicy with `prove`/`onFailure`), EvaluationInput, DecisionRecord,
ReplayBundle, PresentationModel, PublicationReceipt, provider request/response envelope, and
the adopter test-fixture format (ADR-0014) — remembering ADR-0017 §7: the schemas ARE the
public API, Go interfaces stay internal, no v1 SDK. Constraints: EntryRef subjects + matcher
domains (§5), obligations not vouch (§2), `require-review` distinct from `challenge` (§3),
typed fact states with `maxAge` (§4/§6), merge-result pins in preconditions (§1), apiVersion
group settled (OQ-1 residual). **Exit gate (D-016)**: the strict end-to-end contract fixture
of §8 — pinned target/merge result, renamed entry with stable identity, two required
obligations, an expired typed fact, a missing required approval, expected DecisionRecord +
redacted PresentationModel + publication preconditions — validates against the same schemas
as every example; no engine code before it exists. Story seeds: schema-per-contract slices;
fixture authoring; schema-validation CI job; predicate-scope table freeze (ADR-0013);
`assent lint` hard-error list spec (ADR-0010 amendment); OQ-8/9/10/13/21/22 resolutions.

### P3-E2 — Versioning & compatibility spec — Planned
Specify ADR-0017 §9 as testable rules: safety-bearing resources reject unknown
fields/duplicate IDs/unknown enums; reports are additive-tolerant; provider majors
negotiated; named collections are lists with mandatory unique IDs; hashes over canonical
JSON with schema-version domain separation; do-not-generalize list restated as lint errors.
Produces the public `API_STABILITY.md` (oss-playbook #8) stating per-version guarantees for
policy schema / decision contract / test format and graduation criteria. **Exit gate**:
compat rules expressed as executable schema/fixture tests (valid + adversarial invalid
fixtures per rule). Story seeds: strict-decode fixture suite; canonical-JSON hash vectors;
version-negotiation matrix; API_STABILITY.md.

### P3-E3 — Example & pack migration to the frozen contracts — Planned
Migrate `examples/policies/**`, `examples/archetypes/**` (P1-E2 DRAFT fixtures) and the
starter-pack sketches to the frozen schemas: `prove:`/`onFailure:` replacing `effect: vouch`,
matcher domains replacing `path` overloading, DRAFT markers removed (ADR-0017 consequences,
P2-5). **Exit gate**: every committed example validates in CI against the Phase-3 schemas;
the archetype fixtures become the seed corpus for `assent test` goldens. Story seeds:
per-shape starter pack (topic-registry / service-catalog / infra-vars); rego example
quarantined behind a `# locked: D-012` marker; docs updated.

## Phase 4 — Walking skeleton

### P4-E1 — Walking skeleton with trust boundaries from day one — Planned
Thinnest real slice, TDD: CLI in a GitLab CI job on the Spike-B-chosen e2e profile → parses
a one-field YAML change on a generated sample repo → evaluates one `assert` rule proving one
obligation → posts one resolvable thread or approves + SHA-pinned merges → emits the
DecisionRecord report. Trust-boundary behaviours are in the skeleton, not deferred:
policy loaded from target ref only, `.assent/**` MR → `assent-policy` meta-class BLOCK
golden, SHA-guard rejection on target/source movement (merge-result precondition, ADR-0017
§1), protected-pipeline `doctor` precondition check refusing to arm auto-merge, provider-less
run (no token near evaluation). **Exit gate**: the L3 skeleton e2e green + replayable;
determinism double-run gate active in CI; **adoption gate (D-012): one real repository has
run assent on live MRs** — a synthetic fixture does not count. Story seeds: CLI+CI env
adapter; minimal differ (modify only); minimal obligations aggregation; minimal Reconcile
(thread, approve, pinned merge); report artifact; doctor preconditions; e2e harness reuse
from Spike B.

## Phase 5 — Implementation epics (re-derived from ADR-0013/0015/0016/0017 state)

### E1 — Change model & EntryRef classifier — Planned
JSON/YAML/HCL-tfvars adapters → value tree → structural diff with first-class positions,
deletes, opt-in rename fold (never laxer than delete), input resource limits with
opaque→REVIEW fail-safe (ADR-0003 + amendments); `entries:` declarations → stable `EntryRef`
subjects for document/map/keyed-list collections, unkeyed lists rejected at lint; matcher
domains `files`/`values.pointers`/`fileEvents`/`valueChanges` (ADR-0017 §5); environment +
class routing incl. the built-in `assent-policy` meta-class (ADR-0008/0015 §1). **Exit
gate**: golden + property tests over the archetype corpus incl. adversarial near-threshold
rename pairs; billion-laughs/limits corpus fails closed. Seeds: one adapter per format;
differ; rename fold; EntryRef derivation; classifier; limits.

### E2 — Obligations decision engine — Planned
ADR-0007 as reshaped by ADR-0017 §2–4: per-binding `require:` lists; rules `prove` exactly
one named obligation with `onFailure: {effect, code}`; AND-only composition; union of
denies; `require-review` outcome satisfied only by forge-proven eligible approval;
tri-state predicates fail safe; per-firing points against per-binding thresholds (points are
outcome contributions, `score` is not an effect); one-shot arming restrictions for expiring
facts (§4); deterministic order-independent aggregation into DecisionRecord. **Exit gate**:
golden decision tests for every archetype expected-decision pair; determinism double-run;
"broad rule cannot satisfy an obligation it does not name" adversarial golden. Seeds:
obligation registry; aggregation; points; arming logic; DecisionRecord emission.

### E3 — Policy surface: envelope loader, CEL assert backend, lint — Planned
MergePolicy/RulesetBinding/Config loading with strict decode (P3-E2 rules), target-ref-only
loading (ADR-0015 §1); `assert` all/any/not trees with CEL leaves on cel-go per Spike A
decisions (coercion strategy, cost budget, purity env, per-leaf trace); one activation model
shared with `{{ }}` message interpolation, load-time compile errors (ADR-0016 §2);
`assent lint` hard errors (obligation coverage, reserved classes, fail-open restrictions,
tests-per-rule, unkeyed lists, undeclared projections). **Exit gate**: lint catches every
hard-error fixture; all archetype packs load and evaluate to their expected decisions.
Seeds: loader; CEL backend; tree walker + trace; interpolation; lint.

### E4 — GitLab forge adapter: Snapshot/Resolve/Reconcile — Planned
The `Snapshot → Resolve → Reconcile(DesiredReviewState, Preconditions) → PublicationReceipt`
port (ADR-0017 §7) on GitLab: resolvable threads with stable finding keys (idempotent
upsert), approve/merge with source+target SHA and merge-result preconditions failing closed
(§1, ADR-0015 §2), `require-review` evidence via the P1-E3-S02 approval-eligibility chain,
capability flags per tier (gap → never auto-merge), `doctor` typed capability/precondition
report (ADR-0017 §9). **Exit gate**: L2 cassette contract tests + the L3 conformance cases
"target advanced after evaluation → rejected/re-evaluated" and "MR edits own policy →
BLOCK". Seeds: snapshot; thread reconcile; approval evidence; pinned merge; doctor.

### E5 — Provider host + builtins — Planned
Typed provider protocol from Spike C frozen in P3-E1: builtin providers (forge groups,
OIDC/Keycloak, LDAP, ownership file) + HTTP/exec tier; projection-minimized requests;
`resolved|unavailable|invalid|expired` fact states with `observedAt`/`expiresAt`/`maxAge`
as arming preconditions; write-token isolation (ADR-0015 §7) and digest-pinned exec
binaries; controlling facts never fail open. **Exit gate**: isolation harness from Spike C
promoted to CI; state-machine golden per fact state; lint rejects fail-open on controlling
facts. Seeds: host; each builtin as its own slice; HTTP; exec; freshness arming.

### E6 — Adopter test harness `assent test` + examples — Planned
The frozen ADR-0014 fixture format executed by the production pipeline with providers/forge
stubbed: directory cases + inline `cases.yaml`, `--update` golden flow, `--coverage` with
per-rule both-polarity requirement, obligation/finding-code assertions with `exact` safety
default, determinism double-run per case; dogfooded on `examples/` (packs gate themselves in
CI). **Exit gate**: every shipped example pack green under `assent test`; a deliberately
broken pack fails with the expected/actual diff UX. Seeds: runner; expect matcher; update
flow; coverage lint; examples CI job.

### E7 — E2E & conformance infra — Planned (starts alongside E1)
The Spike-B-chosen CI profile productionized + kind for local/demo; sample-repo generator
seeding the `examples/repos/` shapes; the forge conformance suite as the executable port
definition (ADR-0005) incl. all ADR-0015/0017 adversarial cases; determinism gate;
e2e build tags compiled/vetted on every PR (ADR-0017 consequences); CI security gates
(gitleaks, sanitization check). **Exit gate**: conformance suite runs green against the
GitLab adapter in CI on merge to main; every later epic's L3 cases have a home. Seeds:
profile wiring; repo generator; conformance skeleton; determinism job; security jobs.

### E8 — Renderer & presentation — Planned
ADR-0016 tier 0 only (D-012): renderer-owned envelope (markers, escaping, redaction,
clamping outside customizable regions), config knobs (verbosity per env, emoji, collapse
threshold, locale), CEL message rendering from the PresentationModel, `assent render`
against fixtures, default-theme golden markdown snapshots, template lint, `en` locale
catalog. Tiers 1–2 (slots/full templates) stay designed seams. **Exit gate**: render goldens
green; wording changes provably do not break safety tests (ADR-0014 amendment split).
Seeds: PresentationModel consumer; default theme; render command; goldens; locale catalog.

### E9 — Distribution & release — Planned
oss-playbook execution: goreleaser (binaries, brew, curl+checksum), cosign keyless signing +
SLSA provenance + SBOM, git-cliff notes without SHAs, CI hardening (CodeQL, Scorecard,
govulncheck, SHA-pinned actions), mkdocs-material site publishing `docs/` product pages only
(planning/openspec excluded), README per the formula with honest maturity table, VHS demos +
live demo repo (post-Phase-4), GitLab dogfood mirror decision (OQ-2). **Exit gate**: a tagged
release installs via all three channels and verifies signatures; docs site live. Seeds:
release workflow; supply-chain attestations; CI hardening; docs site; demo assets.

### E10 — GitHub adapter + Actions entrypoint — **Locked (D-012)**
Unlocks with a named consumer. Seam kept honest by the P1-E3-S03 dossier (REQUEST_CHANGES +
conversation-resolution parity, merge queue as merge-result pin, base-ref workflow trust)
and by the conformance suite being forge-neutral (E7). No frozen contract until unlock.

### E11 — Rego predicate backend — **Locked (D-012)**
Tier-2 escape hatch inside the same envelope (ADR-0002 v2): violations-shaped modules,
explicit obligation-proof polarity (no implicit "no violation = proof"), OPA capability
sandbox (D-013). Unlocks with a named consumer whose rules exceed the CEL tier-1 ceiling.

### E12 — `serve` webhook mode — **Locked (D-012)**
v1.x wrapper around the identical pipeline; security requirements already fixed (ADR-0015
§6: HMAC verification, event dedup key, idempotent publishing, per-repo tokens, target-ref
policy loading) plus the ADR-0017 §4 justification: the reconciler that can revoke armed
merges when expiring facts control them. Unlocks with a named consumer; also gates OQ-19/20.

### E13 — Remote policy packs — **Locked (D-012)**
`git::` pack sources pinned by commit SHA (tags are mutable), checksum/signature verified,
same target-ref/no-self-modification trust rules as local policy (ADR-0010 amendment,
ADR-0015 §1). Unlocks with a named consumer running a central policy repo (OQ-5).
