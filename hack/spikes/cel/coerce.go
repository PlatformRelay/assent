package celspike

import (
	"errors"
	"fmt"
	"math"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"gopkg.in/yaml.v3"
)

var (
	errNotBool         = errors.New("cel result is not bool")
	errInt64Overflow   = errors.New("numeric value exceeds int64 range")
	errQuotedNumeric   = errors.New("quoted numeric string is not a number")
	errCrossTypeString = errors.New("string may not be compared numerically")
)

// CoercionCase is one row of the executable YAML/HCL→CEL coercion table.
type CoercionCase struct {
	Name       string
	Source     string // "yaml" | "hcl" | "go"
	YAML       string
	HCL        string
	OldKey     string
	NewKey     string
	Expr       string
	WantBool   *bool  // nil → expect typed error
	WantErrSub string // substring of expected error when WantBool is nil
	Note       string
}

// DecodeYAMLMap unmarshals YAML into map[string]any (gopkg.in/yaml.v3 defaults).
func DecodeYAMLMap(src string) (map[string]any, error) {
	var m map[string]any
	if err := yaml.Unmarshal([]byte(src), &m); err != nil {
		return nil, err
	}
	return m, nil
}

// hclSample is a minimal HCL document for number-vs-string cases.
type hclSample struct {
	Partitions *int    `hcl:"partitions,optional"`
	Label      *string `hcl:"label,optional"`
}

// DecodeHCLSample parses a tiny HCL snippet into Go values as HCL would.
func DecodeHCLSample(src string) (map[string]any, error) {
	var doc hclSample
	if err := hclsimple.Decode("spike.hcl", []byte(src), nil, &doc); err != nil {
		return nil, err
	}
	out := map[string]any{}
	if doc.Partitions != nil {
		out["partitions"] = int64(*doc.Partitions)
	}
	if doc.Label != nil {
		out["label"] = *doc.Label
	}
	return out, nil
}

// AsCELNumber prepares a decoded value for numeric CEL comparison under the
// chosen strategy: preserve int/float identity; never silently parse strings.
// Values outside int64 when presented as oversized floats surface as overflow errors.
func AsCELNumber(v any) (any, error) {
	switch n := v.(type) {
	case int:
		return int64(n), nil
	case int64:
		return n, nil
	case uint64:
		if n > math.MaxInt64 {
			return nil, errInt64Overflow
		}
		return int64(n), nil
	case float64:
		// yaml.v3 emits float64 for non-integers and for integers that lost int form.
		if n > float64(math.MaxInt64) || n < float64(math.MinInt64) {
			return nil, fmt.Errorf("%w: %g", errInt64Overflow, n)
		}
		return n, nil
	case string:
		return nil, fmt.Errorf("%w: %q", errQuotedNumeric, n)
	case bool:
		return nil, fmt.Errorf("%w: bool %v", errCrossTypeString, n)
	default:
		return nil, fmt.Errorf("unsupported numeric candidate %T", v)
	}
}

// AsCELBool accepts only real booleans — YAML 1.1 yes/no decode as strings in yaml.v3
// and CEL's == against bool silently returns false (not an error). Adapter must reject.
func AsCELBool(v any) (bool, error) {
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("%w: want bool, got %T (%v)", errCrossTypeString, v, v)
	}
	return b, nil
}
