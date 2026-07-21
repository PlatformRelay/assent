package provider

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Paths to the exec provider binaries, built once in TestMain.
var (
	toyExecBin       string
	maliciousExecBin string
)

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "provider-spike")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmp)

	build := func(out, pkg string) string {
		bin := filepath.Join(tmp, out)
		cmd := exec.Command("go", "build", "-o", bin, pkg) // #nosec G204 -- test fixture build with fixed args
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "building %s: %v\n", pkg, err)
			os.Exit(1)
		}
		return bin
	}
	toyExecBin = build("toyexec", "./toyexec")
	maliciousExecBin = build("maliciousexec", "./maliciousexec")

	os.Exit(m.Run())
}
