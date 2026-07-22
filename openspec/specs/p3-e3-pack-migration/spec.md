# P3-E3 — Example & pack migration to the frozen contracts

**Problem**: `examples/policies/**` and every `examples/archetypes/**` fixture (P1-E2) still
carry "DRAFT — pre-schema" markers and pre-ADR-0017 vocabulary (`effect: vouch`, a bare `path`
key doing double duty as file-glob and value-pointer). Two of them encode the exact bugs
ADR-0017 fixed: the ownership example's "owner must resolve this thread" message treats
self-resolution as authorization (§3's canonical bug), and the no-destruction fixture already
anticipates a `require-review` `onFailure` effect that has no home in the frozen enum P3-E1
has drafted so far (flagged below, not silently resolved). The imagined `assent init --sample`
starter-pack output (`docs/usage/walkthrough.md`) has no committed counterpart at all. Left
alone, none of this content can validate once P3-E1's schemas land (D-016's exit gate), and
adopters trying the three P1-E1 sample shapes have nothing runnable to start from.
**Scope**: migrating the two committed declarative policy examples and every archetype
fixture's `expected.yaml`/`facts.yaml` to `prove`/`onFailure` + the four ADR-0017 §5 matcher
domains, with DRAFT markers removed; authoring one committed starter pack per P1-E1 sample
shape (topic-registry / service-catalog / infra-vars); quarantining the Rego escape-hatch
example; updating the walkthrough + sample READMEs; and registering the migrated archetype
corpus as the seed manifest the future `assent test` golden-run contract must satisfy.
**Non-goals**: authoring or modifying the JSON Schemas themselves (P3-E1 owns every schema
file); implementing `assent test`/`assent lint`/the config loader (Phase 5 engine code, D-016
forbids it before the exit-gate fixture exists); the generic schema-validation CI job itself
(P3-E1-S06 owns it — this epic only adds migration-specific guards that run alongside it);
editing `later-phases.md`, other epics' spec directories, or any ADR — this epic's stories
reference P3-E1/ADR-0017 without changing them.
ADRs: ADR-0017 §2 (required obligations replace anonymous vouch coverage), §3 (challenge is
acknowledgement; `require-review` is authorization — the ownership/no-destruction fix), §5
(governed subjects + the four matcher domains, ending the `path` overload), §9 (strict-decode
posture examples must satisfy); ADR-0017 consequences / P2-5 ("examples migrate to
`prove`/`onFailure` when the schemas land — until then they carry DRAFT markers"); ADR-0003
(opaque change -> REVIEW, referenced by the infra-vars starter pack); ADR-0014 (adopter
test-fixture directory shape, reused by the starter packs' own fixtures); D-012/D-017/E11
(Rego is contract-unlocked but implementation-gated to post-Phase-4 — the quarantine's actual
status); P1-E1 (the three sample repos these packs govern); P1-E2 (`docs/planning/
archetypes.md`, the archetype inventory this epic migrates and later registers as a corpus).

**Cross-epic note (authoring-now, not blocking)**: every story below can be authored and
committed now, per the operator's brief, even though P3-E1's schemas do not exist yet.
Each story's Dependencies/Definition-of-done distinguishes *authoring* this content (no
dependency) from *closing* the story as fully validated (depends on P3-E1's schemas being
integrated and this epic's content passing the P3-E1-S06 CI job) — mirroring the convention
P3-E2/P3-E4 already used for their own forward-referenced dependencies on P3-E1.

## P3-E3-S01 — Migrate the committed declarative policy examples + every archetype fixture to `prove`/`onFailure` + matcher domains

- **Goal**: `examples/policies/declarative/{ownership,bounded-change}.yaml` and every
  `examples/archetypes/**/{expected.yaml,facts.yaml}` (base + `negative/` variants) are
  rewritten to the frozen vocabulary — `prove: {obligation, when}` / `onFailure: {effect,
  code}` instead of `effect: vouch` / `onFail:`; one of the four ADR-0017 §5 matcher domains
  instead of a bare `path` key; the frozen `DecisionRecord` shape (`decision`, `findings:
  [{rule, obligation, effect, subject, points}]`) instead of the ad hoc `firing:
  [{rule_intent, obligation, outcome}]` shape — and no file anywhere under `examples/policies/`
  or `examples/archetypes/` still carries a "DRAFT" marker. The two documented canonical bugs
  (ownership's self-resolvable "owner must resolve" wording; no-destruction's `require-review`
  effect with no enum slot) are fixed or explicitly flagged, never silently carried forward.
- **Operator input**: no.
- **Dependencies**: P3-E1-S01 (the `MergePolicy` schema, matcher domains, and `onFailure`
  effect enum this content must decode against) and P3-E1-S02 (the `DecisionRecord` shape the
  archetype fixtures restate) as *validation* prerequisites — authoring happens now; closing
  this story (not merely authoring it) requires those schemas integrated and this content
  passing the P3-E1-S06 CI job. **Flags one apparent gap in P3-E1-S01-03**: its `onFailure`
  effect enum is `comment | challenge | block` — it omits `require-review`, which ADR-0017 §3
  and the already-accepted P1-E2 inventory (`docs/planning/archetypes.md`) both require for
  ownership's and no-destruction's authorization failure (`block` is semantically wrong there:
  it has "no auto-merge path" at all, whereas an eligible owner's approval must be able to
  unblock these). Logged to the INBOX for P3-E1 to reconcile; this story authors against
  ADR-0017 §3 directly rather than downgrading to an enum value that changes the archetype's
  meaning.
- **Definition of done**: neither committed declarative example contains `effect: vouch`,
  `onFail:`, or a bare `path` key inside `match:`; every archetype `expected.yaml`/`facts.yaml`
  uses `findings`/`decision` (no `firing`/`rule_intent`); no file under `examples/policies/` or
  `examples/archetypes/` contains the string "DRAFT" (the Rego file's DRAFT is replaced by the
  S03 quarantine marker, not deleted); ownership's fix and no-destruction's flagged
  `require-review` usage are both traceable to this story's REQs below.

Requirements:

- **REQ-P3-E3-S01-01** — Given `examples/policies/declarative/ownership.yaml`'s current
  `effect: vouch` / `onFail: {effect: challenge, message: "... owner must resolve ..."}` — the
  exact bug ADR-0017 §3 names by title — when migrated, then the rule becomes `prove:
  {obligation: ownership, when: "entry.owner in facts.author.groups"}` with `onFailure:
  {effect: require-review, code: ownership.unauthorized}`, the "owner must resolve this
  thread" wording is removed, and a comment states authorization is proven only by
  `ApprovalEvidence` (P3-E1-S04) — never by the affected author resolving a thread.
  Adversarial case documented in the file's own comment: an author who is not in the owning
  group must not be able to satisfy this obligation by any forge action they themselves can
  take.
  - Test: `examples/policies/declarative/ownership.yaml`
  - Verify: `! grep -q "effect: vouch" examples/policies/declarative/ownership.yaml && grep -q "prove:" examples/policies/declarative/ownership.yaml && grep -q "require-review" examples/policies/declarative/ownership.yaml`
  - Level: doc
- **REQ-P3-E3-S01-02** — Given `examples/policies/declarative/bounded-change.yaml`'s `match:
  changes: [{path: "**/partitions", kind: modify}]` — the `path`-as-glob-and-pointer overload
  ADR-0017 §5 ends — when migrated, then `match` uses the `valueChanges` matcher domain naming
  the `/partitions` field directly (no glob, no reused `path` key), the two CEL leaves keep
  referencing `old`/`new` unchanged, `effect: vouch` becomes `prove: {obligation:
  bounded-change, when: {all: [...]}}` with `onFailure: {effect: challenge, code:
  bounded-change.out-of-band}` (challenge stays correct here per ADR-0017 §3 — an over-quota
  bump is acknowledgement, not an authorization gap, unlike ownership above).
  - Test: `examples/policies/declarative/bounded-change.yaml`
  - Verify: `! grep -q 'path: "\*\*' examples/policies/declarative/bounded-change.yaml && grep -q "valueChanges" examples/policies/declarative/bounded-change.yaml && grep -q "prove:" examples/policies/declarative/bounded-change.yaml`
  - Level: doc
- **REQ-P3-E3-S01-03** — Given every `examples/archetypes/**/{expected.yaml,facts.yaml}`
  file's ad hoc `firing: [{rule_intent, obligation, outcome, detail}]` shape and "DRAFT —
  pre-schema" header, when migrated, then each `expected.yaml` restates its outcome using the
  frozen `DecisionRecord` shape (P3-E1-S02) — `decision`, `findings: [{rule, obligation,
  effect, subject, points}]` with `subject` as an `EntryRef` string — no `firing`/`rule_intent`
  key remains anywhere, and no file under `examples/policies/` or `examples/archetypes/`
  contains the string "DRAFT" (checked case-insensitively, repo-wide across both directories in
  one pass — not per-file, so a missed file cannot hide).
  - Test: `examples/archetypes/ownership/expected.yaml`
  - Verify: `! grep -rqi "DRAFT" examples/policies examples/archetypes && ! grep -rq "rule_intent" examples/archetypes`
  - Level: doc
- **REQ-P3-E3-S01-04** — Given `examples/archetypes/no-destruction/expected.yaml`'s current
  `# DRAFT — pre-schema` header and its `onFailure: {effect: require-review, code:
  destruction.forbidden}` sitting alongside its delete/rename/near-similarity case index, when
  migrated, then: the DRAFT header is removed from **this file specifically** (not merely
  asserted repo-wide by S01-03 — this REQ's own Verify checks this file by name so a partial
  migration that skips it cannot pass); the file **keeps** `require-review` (per the flagged
  gap above — this is the second archetype the same fix applies to, not a one-off) rather than
  being silently downgraded to `block`, and its existing "never challenge-as-authorization"
  inline comment survives verbatim (proving the effect wasn't merely left alone by accident,
  but re-affirmed after the DRAFT-removal edit touched the same block); the case index's own
  `path:` keys (case-directory names — `delete/`, `rename/`, `near-similarity/` — unrelated to
  the ADR-0017 §5 matcher-domain overload, which only governs a rule's `match:` block) are
  left untouched, but gain a new clarifying comment stating literally that they are "not a
  policy matcher"; and the ADR-0003 "rename never laxer than delete" invariant plus the
  near-similarity adversarial case (`max(strictness(delete), strictness(rename))`) are
  preserved verbatim. Adversarial case for this REQ itself: `require-review`/`near-similarity`
  already appear in the file *before* any migration work — a Verify checking only their
  presence would pass vacuously, so it must also fail while the DRAFT header and the new
  clarifying comment are absent.
  - Test: `examples/archetypes/no-destruction/expected.yaml`
  - Verify: `! grep -qi "DRAFT" examples/archetypes/no-destruction/expected.yaml && grep -qi "not a policy matcher" examples/archetypes/no-destruction/expected.yaml && grep -q "never challenge-as-authorization" examples/archetypes/no-destruction/expected.yaml && grep -q "require-review" examples/archetypes/no-destruction/expected.yaml && grep -q "near-similarity" examples/archetypes/no-destruction/expected.yaml`
  - Level: doc

## P3-E3-S02 — Per-shape starter packs (topic-registry / service-catalog / infra-vars)

- **Goal**: each P1-E1 sample repo shape gets a committed, frozen-schema-shaped starter pack
  under `examples/packs/<shape>/.assent/` — the `assent init --sample <shape>` output the
  walkthrough only imagines becomes real, reviewable example content — covering exactly the
  archetypes that shape's own `examples/repos/<shape>/README.md` already commits to
  exercising, reusing S01's migrated obligations rather than inventing new ones.
- **Operator input**: no.
- **Dependencies**: P1-E1 (the three sample repos these packs govern); P3-E3-S01 (the
  migrated declarative examples/archetype obligations these packs draw rule bodies from);
  P3-E1 (`Config`/`RulesetBinding`/`MergePolicy` schemas) as a validation prerequisite —
  authored now, closed only once P3-E1 lands and each pack validates in the P3-E1-S06 CI job.
- **Definition of done**: three pack bundles committed, one per shape; each pack's rule set is
  a subset of the obligations S01 already migrated/named in the P1-E2 inventory — no new
  obligation invented outside it; each rule has at least one passing and one failing fixture
  under the pack's own `tests/` directory (ADR-0014 shape); each shape's sample-repo README
  gains a one-line pointer to its pack directory (tracked by S04, not this story).

Requirements:

- **REQ-P3-E3-S02-01** — Given `examples/repos/topic-registry/README.md`'s committed
  archetype list (ownership, bounded change, no destruction, environment split, schema
  validity), when the starter pack is authored, then `examples/packs/topic-registry/.assent/
  {config.yaml,bindings.yaml}` classify `topics/<env>/*.yaml` as class `kafka-topic` with
  `entries: {mode: document, identity.pointer: /name}`, bind it to one pack named `topics`
  under `examples/packs/topic-registry/.assent/packs/topics/rules/*.yaml`, and that pack's
  rules `prove` exactly `ownership`, `bounded-change`, `non-destructive`, and `schema-valid` —
  no more, no fewer — each `onFailure` matching the effect S01 already fixed for that
  obligation.
  - Test: `examples/packs/topic-registry/.assent/packs/topics/rules/`
  - Verify: `grep -rq "prove:" examples/packs/topic-registry/.assent/packs/topics/rules/ && grep -rq "ownership" examples/packs/topic-registry/.assent/packs/topics/rules/`
  - Level: doc
- **REQ-P3-E3-S02-02** — Given `examples/repos/service-catalog/README.md`'s archetype list
  (allow-listed fields, ownership, schema validity, no destruction, environment split,
  freshness), when the starter pack is authored, then `examples/packs/service-catalog/.assent/
  packs/catalog/` classifies `catalog/<env>/*.json` entries as a **keyed list** (`entries:
  {mode: list, identity.pointer: /name}` — never unkeyed, per ADR-0017 §5), and its rules
  prove `allowed-fields`, `ownership`, `schema-valid`, and `context-fresh` (the `oncall`
  freshness check), with `non-destructive` covering whole-entry removal from the multi-entry
  file. Adversarial case: a fixture reordering two catalog entries without changing either
  must not change which entry a finding's `subject` (`EntryRef`) resolves to.
  - Test: `examples/packs/service-catalog/.assent/packs/catalog/rules/`
  - Verify: `grep -q "identity" examples/packs/service-catalog/.assent/config.yaml && grep -rq "allowed-fields" examples/packs/service-catalog/.assent/packs/catalog/rules/`
  - Level: doc
- **REQ-P3-E3-S02-03** — Given `examples/repos/infra-vars/README.md`'s archetype list
  (environment split, bounded change, ownership, opaque-change fallback, no destruction), when
  the starter pack is authored, then `examples/packs/infra-vars/.assent/packs/vars/` classifies
  `envs/<env>/*.tfvars` entries with rules proving `bounded-change` (`memory_mb`/
  `min_replicas`/`max_replicas` bands) and `ownership` — and, because no `opaque-change`
  archetype fixture exists yet under `examples/archetypes/` (a P1-E2 gap, not this epic's to
  fill), the pack's `config.yaml` documents the HCL-parse-failure fallback directly against
  ADR-0003 (`opaque -> REVIEW`, never silently vouched/proved) instead of inventing an
  obligation the inventory does not name.
  - Test: `examples/packs/infra-vars/.assent/config.yaml`
  - Verify: `grep -q "opaque" examples/packs/infra-vars/.assent/config.yaml`
  - Level: doc
- **REQ-P3-E3-S02-04** — Given ADR-0014's adopter test-fixture directory shape, when each
  starter pack is authored, then `examples/packs/<shape>/.assent/tests/<pack>/` contains, per
  rule the pack proves, at least one passing fixture (mirrors the matching
  `examples/archetypes/**/base`+`head`+`facts.yaml` triple already migrated in S01) and at
  least one failing fixture (mirrors the matching `.../negative/**` triple) — so `assent test`
  (Phase 5) has a real, runnable-shaped corpus for every starter pack from day one, not merely
  for the archetype inventory considered in isolation.
  - Test: `examples/packs/topic-registry/.assent/tests/topics/`
  - Verify: `find examples/packs -path "*/tests/*/negative" | grep -q packs && find examples/packs -path "*/tests/*/base" | grep -q packs`
  - Level: doc

## P3-E3-S03 — Quarantine the Rego escape-hatch example behind a `# locked: D-012` marker

- **Goal**: `examples/policies/rego/bounded_change.rego` is marked contract-committed-but-
  implementation-gated so neither a reader skimming `examples/` nor a future CI job mistakes
  it for a v1-runnable predicate backend, and it is structurally excluded from S01's migration
  and S02's starter packs (and, later, the `assent test` golden corpus) until E11's
  post-Phase-4 implementation lands.
- **Operator input**: no.
- **Dependencies**: none for authoring the marker convention itself. **Note on the marker's
  literal wording**: the epic brief names `# locked: D-012`, but E11 (the Rego backend) is
  status **"Unlocked (D-017), implementation after Phase 4"** — a milder gate than E10/E13's
  actual `Locked (D-012)`. This story keeps the literal `# locked: D-012` token (for grep
  consistency with the epic's own naming) but requires the accompanying comment to state the
  accurate status, so the marker cannot be misread as "the contract itself is still locked."
- **Definition of done**: the file carries the marker with accurate accompanying prose; a doc
  states what the marker means and who may remove it (E11's own implementation lane, post-
  Phase-4); no example under S01/S02 references or depends on the rego file.

Requirements:

- **REQ-P3-E3-S03-01** — Given the epic brief's instruction to quarantine "behind a `# locked:
  D-012` marker" and E11's actual status line, when the marker is added to `examples/policies/
  rego/bounded_change.rego`, then its first line reads `# locked: D-012 — contract unlocked
  (D-017/E11), implementation gated until the Phase-4 adoption gate; do not wire into assent
  test or the schema-validation CI job before then` (or materially equivalent wording
  preserving both facts) — never wording that implies the rego *contract* itself is still
  D-012-locked.
  - Test: `examples/policies/rego/bounded_change.rego`
  - Verify: `head -1 examples/policies/rego/bounded_change.rego | grep -q "locked: D-012"`
  - Level: doc
- **REQ-P3-E3-S03-02** — Given the marker exists, when the S04 CI guard runs, then it asserts
  every file under `examples/policies/rego/**` carries the marker on or before its first
  non-comment line, and separately asserts no file under `examples/policies/declarative/**`,
  `examples/archetypes/**`, or `examples/packs/**` (S01/S02) references a `rego:` predicate
  leaf — keeping the quarantine structurally enforced, not merely documented. Adversarial
  case for this REQ itself: none of those directories reference a `rego:` leaf *today either*
  (no rego integration exists yet, migrated or not), so a Verify checking only their absence
  would pass vacuously before the marker (S03-01) even exists; it must also require the
  marker's presence, so the check only means something once there is a quarantined escape
  hatch to keep everything else clear of.
  - Test: `examples/policies/rego/bounded_change.rego`
  - Verify: `grep -q "locked: D-012" examples/policies/rego/bounded_change.rego && ! grep -rq "rego:" examples/policies/declarative examples/archetypes examples/packs 2>/dev/null`
  - Level: doc
- **REQ-P3-E3-S03-03** — Given the rego example is excluded from S01's migration, when S01's
  repo-wide DRAFT-marker sweep (REQ-P3-E3-S01-03) runs, then `examples/policies/rego/
  bounded_change.rego`'s own pre-existing "DRAFT" comment is **replaced by** the quarantine
  marker rather than simply deleted — the file must never end up with neither marker (which
  would make it look migrated-and-validated when it is neither).
  - Test: `examples/policies/rego/bounded_change.rego`
  - Verify: `! grep -qi "DRAFT" examples/policies/rego/bounded_change.rego && grep -q "locked: D-012" examples/policies/rego/bounded_change.rego`
  - Level: doc

## P3-E3-S04 — Docs updated; migration-invariant CI guard; archetype corpus registered as the `assent test` seed manifest

- **Goal**: the design-fiction walkthrough and every sample-repo README stop describing
  pre-ADR-0017 vocabulary (`vouch`, a bare `path`, "owner must resolve"); a small, fast,
  migration-specific CI guard (distinct from and running before P3-E1-S06's generic
  schema-validation job) fails a PR that reintroduces a DRAFT marker or an unquarantined
  `rego:` reference so a regression is diagnosed by name, not by an opaque schema error; and
  the migrated archetype fixtures are registered as the named, versioned seed corpus the
  future `assent test` golden-run contract must satisfy, per the epic's exit gate.
- **Operator input**: no.
- **Dependencies**: P3-E3-S01/S02/S03 (nothing to point docs at or guard in CI until the
  migration exists); P3-E1-S06 (the generic schema-validation CI job this story's guard
  supplements — never duplicates: S06 validates schema conformance by `apiVersion`/`kind`,
  this story only checks the migration-specific invariants a schema cannot express, like
  marker text).
- **Definition of done**: `docs/usage/walkthrough.md` and all three `examples/repos/*/
  README.md` files read consistently with `prove`/`onFailure`/matcher-domain vocabulary; a CI
  guard script exists and fails on each of the three adversarial cases named in its own REQ;
  `docs/planning/archetype-goldens.md` lists every archetype directory and its expected
  `assent test` outcome as a versioned manifest.

Requirements:

- **REQ-P3-E3-S04-01** — Given `docs/usage/walkthrough.md`'s `effect: vouch`-era vocabulary
  ("vouched", a bare `assert "new < old"` shown without any obligation/prove/onFailure
  framing), when updated, then every command-output snippet and prose reference uses
  `prove`/`onFailure`/obligation vocabulary consistent with S01's migrated examples, the file's
  own "design fiction" disclaimer is kept verbatim (this remains speculative CLI UX, not a
  frozen contract), and no line presents a bare `path` glob as the only matcher shape a rule
  can declare.
  - Test: `docs/usage/walkthrough.md`
  - Verify: `! grep -q "vouched" docs/usage/walkthrough.md && grep -q "prove" docs/usage/walkthrough.md && grep -q "design fiction" docs/usage/walkthrough.md`
  - Level: doc
- **REQ-P3-E3-S04-02** — Given each `examples/repos/{topic-registry,service-catalog,
  infra-vars}/README.md`'s existing "Rule archetypes exercised" section, when updated, then
  each gains one line pointing at its S02 starter pack (`examples/packs/<shape>/`) so a reader
  lands on runnable policy content, not just a prose archetype list.
  - Test: `examples/repos/topic-registry/README.md`
  - Verify: `grep -q "examples/packs/topic-registry" examples/repos/topic-registry/README.md`
  - Level: doc
- **REQ-P3-E3-S04-03** — Given P3-E1-S06 already wires the generic "validate every fixture
  under `examples/**` against its schema" CI job, when this epic's own guard is added (as a
  preceding step in the same workflow, or a standalone `hack/check-migration-invariants.sh`,
  mirroring the `hack/check-sanitization.sh` precedent), then it: (a) fails if any file under
  `examples/policies/**` or `examples/archetypes/**` contains a case-insensitive "DRAFT"
  marker; (b) fails if any file under `examples/policies/rego/**` is missing the S03
  quarantine marker on or before its first non-comment line; and (c) fails if any file outside
  `examples/policies/rego/**` contains a `rego:` predicate leaf — and it runs before the
  P3-E1-S06 schema-validation step in the same job so a migration regression surfaces with a
  named, human-readable failure first.
  - Test: `hack/check-migration-invariants.sh`
  - Verify: `bash hack/check-migration-invariants.sh`
  - Level: L2
- **REQ-P3-E3-S04-04** — Given the epic's exit gate ("archetype fixtures become the seed
  corpus for `assent test` goldens"), when the corpus is registered, then
  `docs/planning/archetype-goldens.md` lists every directory under `examples/archetypes/**`
  that carries a `base`/`head`/`facts.yaml`/`expected.yaml` quadruple (or a `negative/`
  variant of it), states the `decision` each must produce, and is itself versioned so adding
  an archetype later is a reviewable diff against a named manifest — giving the future
  `assent test` implementation (Phase 5) one authoritative list to iterate rather than a
  directory glob it invents itself.
  - Test: `docs/planning/archetype-goldens.md`
  - Verify: `grep -q "ownership" docs/planning/archetype-goldens.md && grep -q "no-destruction" docs/planning/archetype-goldens.md`
  - Level: doc
