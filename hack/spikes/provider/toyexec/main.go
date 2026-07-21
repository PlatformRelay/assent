// Command toyexec is the exec transport of the toy group-membership provider:
// FactQuery on stdin, FactResponse on stdout.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	provider "github.com/PlatformRelay/assent/hack/spikes/provider"
)

func main() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var q provider.FactQuery
	if err := json.Unmarshal(raw, &q); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(provider.ToyAnswer(q)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
