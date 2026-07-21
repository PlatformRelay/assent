# Specs

Spec-driven backlog. Start at **[backlog.md](backlog.md)** — the phase → epic index.

- Phases 1–2 (requirements harvest, spikes & ADR firming): full epics with INVEST stories
  and REQ IDs (`REQ-<epic>-S<story>-<nn>`, each with Given/When/Then, `Test:`, `Verify:`),
  one directory per epic.
- Phases 3–5: epic paragraphs in [later-phases.md](later-phases.md); their stories are
  authored during Phase 3 ("contracts first"), after the contract fixture (ADR-0017 §8)
  exists.

Authoring rules: [../config.yaml](../config.yaml). Levels: L0–L3 per
[ADR-0006](../../docs/adr/0006-testing-strategy.md); design artifacts use `Level: doc`.
