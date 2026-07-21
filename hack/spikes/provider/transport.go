package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// secretName flags environment variable names that must never pass through to
// a provider process, even when explicitly configured (ADR-0015 §7).
var secretName = regexp.MustCompile(`(?i)(TOKEN|SECRET)`)

// ScrubEnv builds a provider environment from scratch: the host process
// environment (which holds the forge write token) is never inherited. Only
// PATH plus explicitly configured, non-credential-looking entries survive.
func ScrubEnv(configured []string) []string {
	env := []string{"PATH=" + os.Getenv("PATH")}
	for _, kv := range configured {
		name, _, _ := strings.Cut(kv, "=")
		if secretName.MatchString(name) {
			continue
		}
		env = append(env, kv)
	}
	return env
}

// CallHTTP posts the FactQuery to an HTTP provider and returns the raw body.
func CallHTTP(ctx context.Context, url string, q FactQuery, timeout time.Duration) ([]byte, error) {
	body, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(resp.Body)
}

// CallExec runs an exec provider with the FactQuery on stdin and a scrubbed
// environment, returning its raw stdout.
func CallExec(ctx context.Context, bin string, configuredEnv []string, q FactQuery, timeout time.Duration) ([]byte, error) {
	body, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	// #nosec G204 -- spike harness: bin is the operator-pinned provider binary, no user input.
	cmd := exec.CommandContext(ctx, bin)
	cmd.Env = ScrubEnv(configuredEnv)
	cmd.Stdin = bytes.NewReader(body)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return out.Bytes(), err
	}
	return out.Bytes(), nil
}
