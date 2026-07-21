package celspike_test

import (
	"strings"
	"testing"

	celspike "github.com/PlatformRelay/assent/hack/spikes/cel"
)

func TestInterpolation(t *testing.T) {
	env, err := celspike.NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	t.Run("valid-template-compiles", func(t *testing.T) {
		tmpl := "quota {{ facts.quota.max_partitions }} exceeded"
		if err := celspike.CompileMessage(env, tmpl); err != nil {
			t.Fatalf("CompileMessage: %v", err)
		}
	})

	t.Run("typo-rejected-at-compile-with-position", func(t *testing.T) {
		tmpl := "quota {{ facts.qota.max_partitions }} exceeded"
		err := celspike.CompileMessage(env, tmpl)
		if err == nil {
			t.Fatal("expected compile error for facts.qota")
		}
		msg := err.Error()
		if strings.Contains(msg, "<no value>") {
			t.Fatalf("must never surface <no value>: %s", msg)
		}
		if !strings.Contains(msg, "message template:") {
			t.Fatalf("want positioned message template error, got %s", msg)
		}
		// Position should point at the slot (line 1, column of '{{').
		if !strings.Contains(msg, "undefined field") && !strings.Contains(msg, "qota") {
			t.Fatalf("want undefined field / qota in error, got %s", msg)
		}
		if !strings.Contains(msg, ":") {
			t.Fatalf("want line:col in error, got %s", msg)
		}
	})
}
