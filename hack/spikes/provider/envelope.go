// Package provider is Spike C (P2-E3): a throwaway proof of the typed,
// minimized HTTP/exec provider contract (ADR-0017 §6) and the token-isolation
// invariant (ADR-0015 §7). Not product code; the schemas and findings feed the
// Phase-3 contract fixture via docs/planning/spikes/spike-c-provider.md.
package provider

import "time"

// Envelope constants: every serialized form carries apiVersion + kind.
const (
	APIVersion       = "provider.assent.dev/v1alpha1"
	KindFactQuery    = "FactQuery"
	KindFactResponse = "FactResponse"
)

// Fact states — distinct machine states, never a silently absent key.
const (
	StateResolved    = "resolved"
	StateUnavailable = "unavailable"
	StateInvalid     = "invalid"
	StateExpired     = "expired"
)

// Subject identifies who or what a fact is about.
type Subject struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// ValueProjection is one declared JSON-Pointer slice of the change — the only
// change content a provider may see.
type ValueProjection struct {
	Pointer string `json:"pointer"`
	Old     any    `json:"old,omitempty"`
	New     any    `json:"new,omitempty"`
}

// Projections groups the declared change slices carried by a query.
type Projections struct {
	Values []ValueProjection `json:"values,omitempty"`
}

// FactQuery is the request envelope the host sends to a provider.
type FactQuery struct {
	APIVersion  string      `json:"apiVersion"`
	Kind        string      `json:"kind"`
	QueryID     string      `json:"queryId"`
	AsOf        time.Time   `json:"asOf"`
	Subject     Subject     `json:"subject"`
	Outputs     []string    `json:"outputs"`
	Projections Projections `json:"projections"`
	DeadlineMs  int         `json:"deadlineMs,omitempty"`
}

// Declaration is the typed output declaration a fact was produced under.
type Declaration struct {
	Type        string `json:"type"`
	Cardinality string `json:"cardinality"`
	Subject     string `json:"subject"`
	Sensitive   bool   `json:"sensitive"`
	MaxAge      string `json:"maxAge"`
}

// Fact is one typed output with an explicit state — never a silently
// absent key.
type Fact struct {
	Name        string      `json:"name"`
	Declaration Declaration `json:"declaration"`
	State       string      `json:"state"`
	Subject     Subject     `json:"subject"`
	ObservedAt  time.Time   `json:"observedAt"`
	ExpiresAt   *time.Time  `json:"expiresAt,omitempty"`
	Value       any         `json:"value,omitempty"`
	Reason      string      `json:"reason,omitempty"`
}

// FactResponse is the response envelope: one fact per requested output.
type FactResponse struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	QueryID    string `json:"queryId"`
	Facts      []Fact `json:"facts"`
}
