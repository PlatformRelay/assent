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
