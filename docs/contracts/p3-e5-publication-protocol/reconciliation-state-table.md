# Reconciliation state table + numbered protocol

**Status**: frozen (P3-E5-S02). **Scope**: this document freezes the nine-step reconciliation
protocol and the state table that the Reconcile port (ADR-0017 §7
`Reconcile(DesiredReviewState, Preconditions) -> PublicationReceipt`) and the finding-lifecycle
state machine (ADR-0011 amendment 2 `UpsertComment`/`SyncThreads`, ADR-0012 amendment 2) already
assume, so that "rerun idempotence" (the P4-E1 exit gate) has something precise to implement
against. It builds on the four marker concepts (`slot`/`occurrence`/`decision`/`artifact`) frozen
in [`marker-grammar.md`](marker-grammar.md) — read that document first for vocabulary. The
duplicate-repair rule, the one-publisher-per-MR topology, and the `doctor` guarantee-reporting
field are **out of scope here** — they are P3-E5-S03's contract (a sibling section or file under
this same directory). Non-goals: authoring ADR-0019 itself; implementing the Reconcile port,
`assent doctor`, or `serve`'s keyed lock — see the epic header in
[`openspec/specs/p3-e5-publication-protocol/spec.md`](../../../openspec/specs/p3-e5-publication-protocol/spec.md).

## The nine-step protocol

Every reconciliation run — a fresh run, a plain rerun, or a rerun after a mid-run crash —
executes these nine steps, in this order, every time. No step is conditional on "did the last run
finish": the protocol is idempotent by construction, not by detecting its own prior completion.

1. **Recompute `DesiredReviewState` from trusted inputs.** The desired state is derived only from
   the `Preconditions` the Reconcile port pins (source/target/merge digests, decision hash, fact
   validity deadline — ADR-0017 §7) — never from a stale local checkout or a cached prior run.
   Invariant protected: **target-ref trust**. A desired state computed from anything other than
   the pinned target ref could authorize publishing findings for content that was never actually
   evaluated.
2. **List paginated bot-authored artifacts.** Every comment/thread the bot identity has posted on
   the MR is enumerated across **all** pages, not just the first — this list is reconciliation's
   only view of "what already exists". Invariant protected: **no-database/D-007**. The forge's
   own comment/thread list is the sole durable record; a truncated (unpaginated) listing silently
   reintroduces exactly the missing-state problem a database would have solved, and is
   indistinguishable from data loss.
3. **Update the one summary slot in place.** The per-MR summary (`artifact.kind:
   summary-comment`) is edited in place; it is never re-posted (ADR-0012 amendment 2). Invariant
   protected: **determinism**. Exactly one summary artifact must exist per MR at all times, so a
   rerun's summary update is a pure function of the current desired state, not of how many prior
   runs posted a summary.
4. **Leave the same unresolved occurrence untouched.** When an existing artifact's `occurrence`
   matches the freshly computed occurrence for its slot, and the artifact's forge thread is still
   unresolved, reconciliation performs **no write** against that artifact. Invariant protected:
   **determinism**. A no-op is itself the correct, repeatable outcome — editing or re-posting an
   already-correct artifact would make the observable result depend on how many times the run has
   executed, not just on the current desired state.
5. **Preserve resolution of the same occurrence across reruns.** When an existing artifact's
   `occurrence` matches the freshly computed occurrence, and the forge's own thread-resolution API
   reports it resolved, reconciliation leaves that resolution alone — it is never re-opened,
   re-derived, or overridden by anything computed in-process. Invariant protected:
   **no-database/D-007**. The forge's resolved/unresolved bit is the only durable record of "was
   this actually reviewed"; reconciliation reads it, it never maintains a competing copy.
6. **Supersede stale occurrences with a fresh challenge.** When an existing artifact's `occurrence`
   no longer matches the freshly computed occurrence for its slot (the judged content changed —
   see `marker-grammar.md`'s occurrence-supersession example), reconciliation posts a **new**
   challenge thread for the new occurrence and leaves the old, now-stale thread resolved-but-stale
   rather than deleting or overwriting it. Invariant protected: **determinism**. A resolved
   "shrink to 7 days — sure?" thread must never be silently reinterpreted as authorizing a later,
   unreviewed shrink to 1 day; the new occurrence gets its own, freshly unresolved review record.
7. **Resolve no-longer-desired findings.** When an existing bot-authored artifact's slot is absent
   from the freshly computed `DesiredReviewState` (the finding no longer fires — e.g. the
   offending change was reverted), reconciliation resolves that artifact with a note ("outdated as
   of `<sha>`") rather than leaving it open. Invariant protected: **target-ref trust**. Because the
   desired state was recomputed from the current pinned target ref (step 1), an artifact missing
   from it is authoritatively no-longer-desired, not merely "not seen this run" — resolving it is
   safe precisely because step 1's trust boundary holds.
8. **Deterministically repair pre-existing duplicates.** When two or more bot-authored artifacts
   are found occupying the same slot (seeded by a race, a prior bug, or a crash — never expected
   from a correctly serialized publisher), reconciliation repairs them by a fixed rule (lowest
   forge ID canonical) rather than first-seen-by-scan-order. The full repair rule, its
   `PublicationReceipt.repairs` shape, and its fixture are **P3-E5-S03's contract**; this step is
   listed here only to fix its position in the sequence. Invariant protected: **determinism**. A
   repair rule that depended on scan order would make the outcome of a duplicate incident
   dependent on pagination timing rather than a fixed, replayable tiebreak.
9. **Rescan after publication before reporting success.** After every write in steps 3–8 has been
   issued, reconciliation re-lists bot-authored artifacts (repeating step 2's listing, not
   trusting its own in-memory record of what it just wrote) and only reports success once that
   rescan confirms the forge itself reflects the desired state. Invariant protected:
   **no-database/D-007**. The forge is the source of truth, not the publisher's in-process plan;
   a crash between issuing a write and this rescan is exactly what step 2 of the *next* run is
   designed to discover and reconcile, so a success report before the rescan would be an
   unverified claim.

## State table — (existing-artifact-state x desired-state) -> action

Every combination of "what already exists for a slot" and "what the current run desires for that
slot" resolves to **exactly one** action below — no row triggers more than one action, and no
combination is left unhandled. The five existing-artifact-state categories are the ones the epic
text names; `desired` / `not desired` is whether the slot appears in the freshly computed
`DesiredReviewState` (step 1).

| # | Existing artifact state | Desired state | Action | Protocol step |
| --- | --- | --- | --- | --- |
| 1 | No artifact exists for this slot | Desired | `create` | step 3 (summary) / step 6-adjacent create path for a first-seen finding |
| 2 | Artifact matches current occurrence, **unresolved** | Desired | `leave-untouched` | step 4 |
| 3 | Artifact matches current occurrence, **resolved** | Desired | `preserve-resolution` | step 5 |
| 4 | Artifact has a **stale occurrence** (differs from the freshly computed occurrence) | Desired (new occurrence) | `supersede-with-fresh-challenge` | step 6 |
| 5 | Artifact exists for this slot | **No-longer-desired** (slot absent from `DesiredReviewState`) | `resolve` | step 7 |

Each row is mutually exclusive by construction: a slot's existing-artifact-state (no artifact /
matches-unresolved / matches-resolved / stale-occurrence / no-longer-desired) is a single
classification per run, computed once per slot from the step-1 desired state and the step-2
artifact listing — so a slot can never simultaneously satisfy two rows and never triggers more
than one action. Row 5 ("no-longer-desired") and row 4 ("stale occurrence") are deliberately
distinct: a no-longer-desired slot has **no** freshly computed occurrence to compare against (the
slot itself dropped out of `DesiredReviewState`), while a stale-occurrence slot is still desired,
just under a different occurrence — conflating the two would resolve findings that should instead
receive a fresh challenge, or supersede findings that should instead be closed outright.

Pre-existing duplicates (two-or-more artifacts occupying one slot) are not a sixth row of this
table — they are a precondition-violation the table assumes step 8 has already repaired down to
one canonical artifact per slot before rows 1–5 are evaluated. See P3-E5-S03 for the repair rule
itself.

## Fixtures

Two fixtures under [`fixtures/`](fixtures/) exercise this protocol end to end and are the
exit-gate artifact P4-E1 consumes for its rerun-idempotence case:

- [`fixtures/rerun-idempotence.yaml`](fixtures/rerun-idempotence.yaml) — a plain rerun (no source
  change between run 1 and run 2) exercising rows 2 and 3 of the state table: run 2 creates zero
  new artifacts and zero duplicate slot occupancy.
- [`fixtures/crash-then-rerun.yaml`](fixtures/crash-then-rerun.yaml) — a run that crashes after
  step 3 but before step 9, exercising step 2's listing as the mechanism that makes the following
  rerun idempotent: run 2 creates nothing new for slots run 1 already covered, and completes the
  rescan run 1 never reached.
