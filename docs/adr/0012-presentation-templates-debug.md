# ADR-0012: Presentation: comment rendering, expandable details, docs links, rule debug

| | |
| --- | --- |
| **Status** | Proposed |
| **Date** | 2026-07-21 |
| **Deciders** | Konrad Heimel |
| **Context links** | [ADR-0007 effects](0007-rule-effects-decision-aggregation.md) · [ADR-0010 config](0010-config-files-repo-layout.md) |

## Context

The comment/thread text is the product's face; a finding the contributor doesn't understand
is a support ticket. Wanted: **best-possible defaults** (zero templating required), full
extendability (repos can rebrand/reformat everything), per-rule documentation links, and
collapsible detail/debug sections so comments stay short but never opaque.

## Decision (proposed)

### Rendering pipeline

`Finding` (structured, ADR-0007) → **renderer** → forge markdown. Rules never emit final
markdown; they emit data + optional template fragments. The renderer owns layout, so all
comments look consistent and dry-run/CI/serve render identically.

### Default layout (per finding)

```markdown
**⚠️ Retention shrinks from 604800000 to 86400000 — data loss possible. Sure?**

Resolve this thread to confirm. `topics/prod/orders.yaml` · rule `topic-safety/retention-shrink-challenge`

<details><summary>Why this check exists & how to fix</summary>

{{ rule docs body or excerpt }}
📖 [Full documentation]({{ rule.docsUrl }})
</details>

<details><summary>Evaluation details</summary>

- matched change: `/retentionMs` modify `604800000 -> 86400000`
- facts used: `quota.max_partitions=24`, `author.groups=[team-orders]`
- score contribution: +0 · pack `topics@a1b2c3d` · engine v0.3.0
</details>
```

Plus one **summary comment** per MR (decision, score vs threshold, finding index) — edited
in place on re-runs, never re-posted.

### Rule-level fields (envelope, ADR-0010)

```yaml
- name: retention-shrink-challenge
  ...
  message: "Retention shrinks from {{ old }} to {{ new }} — data loss possible. Sure?"
  docs:
    url: https://example.com/policies/retention   # link in the collapsible
    summary: "Shrinking retention deletes data irreversibly once segments roll."
  debug:                                          # extra lines for the details section
    - "current consumers: {{ facts.consumers | join ', ' }}"
```

### Extendability

Repo-level template overrides in `.assent/templates/` (finding, summary, decision footers) —
Go `text/template` over the exported Finding/Decision structs, same data contract as the JSON
report. Defaults ship embedded; overriding is opt-in per template, not all-or-nothing.

### Debug & statistics (no DB)

- `explain` mode (ADR-0009) prints the full Trace; `run` embeds it in the JSON report only.
- **`stats`** subcommand aggregates a directory/glob of JSON reports (CI artifacts) into
  automerge rate, outcome distribution, top firing rules, score histograms — flat files in,
  table/JSON out. A database is explicitly out of scope for now; the report artifact is the
  storage format.

## Consequences

- Because rendering is centralized, changing the house style never touches rules or packs.
- `docs.url` becomes a lint warning when missing on `challenge`/`block` rules — those are
  exactly the findings that interrupt humans.

## Counterpoints considered

- *"Let rules write markdown directly."* — Fast at first, then every pack invents its own
  look, templates leak forge-specific syntax, and dry-run output diverges from posted output.

## Amendment (2026-07-21, adversarial review F8): rendering is injection-safe

All interpolated values (`{{ old }}`, `{{ new }}`, facts, file content) are **escaped for
markdown/HTML** and rendered in a **single pass** — user-controlled content is never
re-evaluated as a template, and raw HTML from values is neutralized so authors cannot forge
"approved" banners, close `<details>` blocks early, or hide findings. Untrusted values are
additionally length-clamped in comments (full values live in the JSON report). State is never
parsed back from comment text: thread resolution and decision state always come from the
forge API and the report artifact.

## Amendment 2 (2026-07-21, second review P1-5 / security review A-08)

**Finding lifecycle (state machine, frozen with Phase 3 contracts).** Stable finding key =
`(rule id, file, path, value-hash)`. On every run the publisher reconciles posted state
against current findings via HTML marker comments (decision hash + finding key):
unchanged → leave; new → post; no longer firing → resolve-with-note ("outdated as of
<sha>"); **fired again with a different value-hash → the old resolved thread is stale
consent: post a fresh challenge thread and note the supersession** (a resolved
"retention shrink to X — sure?" never authorizes a later shrink to Y). Re-runs and
crash-recovery are idempotent upserts — never duplicate spam. The summary comment embeds
the decision hash and report-artifact link, and is the only edited-in-place comment.
High-cardinality findings (same rule, many paths) collapse into one thread with a path list
beyond a configurable cap.

**Secret redaction (A-08).** Facts carry a `sensitive` marker (provider- or config-
declared); sensitive values are redacted from comments, debug sections, traces, logs, and
the report artifact by default (report keeps a salted hash for replay comparison).
Renderer additionally scrubs known secret patterns. Providers must never return raw
credentials as facts. Template scope includes `env` so packs can vary wording per
environment honestly (nonprod findings should not claim production stakes).

## Superseded in part (2026-07-21)

The override mechanism of this ADR ("repo-level template overrides … Go `text/template`")
is superseded by [ADR-0016](0016-presentation-theming.md): tiered customization (config
knobs → slots → full templates, adoption-gated), CEL-based `{{ }}` interpolation, a
renderer-owned envelope that user content cannot touch, and the PresentationModel as the
pinned render contract. The default layout, finding-lifecycle state machine, and redaction
rules in this ADR remain in force.
