# DRAFT — illustrative Rego against the draft PolicyInput schema (frozen in Phase 3).
#
# Archetype "bounded change": partitions may only increase, and only up to quota.
# Any violation yields a finding; findings of severity "review" force human review.
package policies.bounded_change

import rego.v1

# input.changes: [{path, file, kind, old, new}]  (canonical ChangeSet, ADR-0003)
# input.facts:   provider results, e.g. quotas   (ADR-0004)

deny contains finding if {
	some change in input.changes
	change.kind == "modify"
	endswith(change.path, "/partitions")
	change.new < change.old
	finding := {
		"rule": "bounded-change/no-decrease",
		"severity": "review",
		"path": change.path,
		"message": sprintf("partitions may not decrease (%d -> %d)", [change.old, change.new]),
	}
}

deny contains finding if {
	some change in input.changes
	change.kind == "modify"
	endswith(change.path, "/partitions")
	change.new > input.facts.quota.max_partitions
	finding := {
		"rule": "bounded-change/quota",
		"severity": "review",
		"path": change.path,
		"message": sprintf("partitions %d exceeds quota %d", [change.new, input.facts.quota.max_partitions]),
	}
}
