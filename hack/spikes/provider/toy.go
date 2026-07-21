package provider

import (
	"encoding/json"
	"net/http"
	"time"
)

// Toy group-membership fixture — deterministic, keyed only by the query.
var toyGroups = map[string][]string{
	"alice": {"platform-team", "reviewers"},
	"bob":   {"reviewers"},
}

const toyMaxAge = time.Hour

func groupsDeclaration() Declaration {
	return Declaration{
		Type:        "string",
		Cardinality: "set",
		Subject:     "user",
		Sensitive:   false,
		MaxAge:      "1h",
	}
}

// ToyAnswer is the single group-membership implementation shared by both
// transports: identical queries must yield identical envelopes. All timestamps
// derive from the host-pinned q.AsOf — no wall clock.
func ToyAnswer(q FactQuery) FactResponse {
	facts := make([]Fact, 0, len(q.Outputs))
	for _, name := range q.Outputs {
		fact := Fact{
			Name:        name,
			Declaration: groupsDeclaration(),
			Subject:     q.Subject,
			ObservedAt:  q.AsOf,
		}
		if name != "groups" {
			fact.State = StateInvalid
			fact.Reason = "output not declared by this provider"
		} else {
			groups := toyGroups[q.Subject.ID]
			if groups == nil {
				groups = []string{}
			}
			expires := q.AsOf.Add(toyMaxAge)
			fact.State = StateResolved
			fact.Value = groups
			fact.ExpiresAt = &expires
		}
		facts = append(facts, fact)
	}
	return FactResponse{
		APIVersion: APIVersion,
		Kind:       KindFactResponse,
		QueryID:    q.QueryID,
		Facts:      facts,
	}
}

// ToyHTTPHandler serves ToyAnswer as the HTTP transport of the same provider.
func ToyHTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var q FactQuery
		if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ToyAnswer(q))
	})
}
