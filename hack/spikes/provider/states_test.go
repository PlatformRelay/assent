package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// REQ-P2-E3-S01-02: timeout → unavailable, garbage → invalid, stale expiresAt
// → expired. Distinct machine states, never a silently absent key, and never
// resolved — controlling facts fail closed.
func TestStates(t *testing.T) {
	q := groupQuery(t)
	// Host clock is injected: everything is relative to the pinned asOf.
	now := fixedAsOf

	// The timeout handler blocks on a test-owned channel (released after the
	// harness call) so srv.Close never waits on a live connection.
	release := make(chan struct{})

	handlers := map[string]http.Handler{
		"timeout": http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			select {
			case <-release:
			case <-r.Context().Done():
			}
		}),
		"garbage": http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("{{{ not json"))
		}),
		"stale": http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			resp := ToyAnswer(q)
			stale := q.AsOf.Add(-time.Minute)
			for i := range resp.Facts {
				resp.Facts[i].ExpiresAt = &stale
			}
			writeJSON(t, w, resp)
		}),
	}

	wantState := map[string]string{
		"timeout": StateUnavailable,
		"garbage": StateInvalid,
		"stale":   StateExpired,
	}

	got := map[string]string{}
	for name, h := range handlers {
		srv := httptest.NewServer(h)
		call := func(ctx context.Context) ([]byte, error) {
			return CallHTTP(ctx, srv.URL, q, 100*time.Millisecond)
		}
		facts := ResolveFacts(t.Context(), call, q, now)
		if name == "timeout" {
			close(release)
		}
		srv.Close()

		fact, ok := facts["groups"]
		if !ok {
			t.Fatalf("%s: fact key silently absent — must never happen", name)
		}
		if fact.State != wantState[name] {
			t.Errorf("%s: state = %q, want %q (reason: %q)", name, fact.State, wantState[name], fact.Reason)
		}
		if fact.State == StateResolved {
			t.Errorf("%s: controlling fact resolved on a failure path — must fail closed", name)
		}
		got[name] = fact.State
	}

	if got["timeout"] == got["garbage"] || got["garbage"] == got["stale"] || got["timeout"] == got["stale"] {
		t.Errorf("failure states are not distinct: %v", got)
	}
}
