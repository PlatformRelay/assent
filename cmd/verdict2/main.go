// Command verdict2 is the CLI entry point for the auto-merge gate.
//
// Pre-alpha stub: real subcommands (run, test, lint, render) arrive with the
// walking skeleton (meta-plan Phase 4). Kept minimal so `go build ./...` and
// CI gates are meaningful from day one.
package main

import (
	"fmt"
	"os"
)

var version = "0.0.0-dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("verdict2 " + version)
		return
	}
	fmt.Fprintln(os.Stderr, "verdict2 (pre-alpha): no commands implemented yet — see docs/planning/meta-plan.md")
	os.Exit(2)
}
