package provider

import (
	"context"
	"encoding/json"
	"time"
)

// CallFunc is one transport invocation (HTTP or exec) for a FactQuery.
type CallFunc func(ctx context.Context) ([]byte, error)

// ResolveFacts is the host side of the protocol: it calls the provider and
// classifies the outcome into exactly one Fact per requested output. Every
// failure path lands in a distinct non-resolved state — a controlling fact is
// therefore fail-closed by construction and never a silently absent key.
// `now` is injected by the host (the pinned evaluation instant), never read
// from a wall clock here.
func ResolveFacts(ctx context.Context, call CallFunc, q FactQuery, now time.Time) map[string]Fact {
	raw, err := call(ctx)
	if err != nil {
		return synthesizeAll(q, StateUnavailable, "provider call failed: "+err.Error(), now)
	}
	if err := ValidateResponse(raw); err != nil {
		return synthesizeAll(q, StateInvalid, "response failed schema validation", now)
	}
	var resp FactResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return synthesizeAll(q, StateInvalid, "response undecodable", now)
	}
	if resp.QueryID != q.QueryID {
		return synthesizeAll(q, StateInvalid, "queryId mismatch", now)
	}

	byName := make(map[string]Fact, len(resp.Facts))
	for _, f := range resp.Facts {
		byName[f.Name] = f
	}

	out := make(map[string]Fact, len(q.Outputs))
	for _, name := range q.Outputs {
		fact, ok := byName[name]
		if !ok {
			out[name] = synthesize(q, name, StateInvalid, "provider omitted a requested output", now)
			continue
		}
		if fact.State == StateResolved && fact.ExpiresAt != nil && !fact.ExpiresAt.After(now) {
			fact.State = StateExpired
			fact.Reason = "expiresAt is not after the evaluation instant"
			fact.Value = nil
		}
		out[name] = fact
	}
	return out
}

func synthesizeAll(q FactQuery, state, reason string, now time.Time) map[string]Fact {
	out := make(map[string]Fact, len(q.Outputs))
	for _, name := range q.Outputs {
		out[name] = synthesize(q, name, state, reason, now)
	}
	return out
}

func synthesize(q FactQuery, name, state, reason string, now time.Time) Fact {
	return Fact{
		Name:       name,
		State:      state,
		Subject:    q.Subject,
		ObservedAt: now,
		Reason:     reason,
	}
}
