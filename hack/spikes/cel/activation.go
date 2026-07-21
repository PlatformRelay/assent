// Package celspike is a throwaway Spike A harness for CEL residual risks (ADR-0013).
// Excluded from the D-010 coverage gate; tests are real and runnable.
package celspike

import (
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
)

// QuotaFacts mirrors examples/archetypes/*/facts.yaml quota block.
type QuotaFacts struct {
	MaxPartitions int64 `cel:"max_partitions"`
}

// AuthorFacts mirrors facts.author.
type AuthorFacts struct {
	Login  string   `cel:"login"`
	Groups []string `cel:"groups"`
}

// Facts is the typed activation fragment for compile-time field checks (ADR-0016 §2).
type Facts struct {
	Quota  QuotaFacts  `cel:"quota"`
	Author AuthorFacts `cel:"author"`
}

// CostBudget is the recommended program cost limit measured in this spike.
// Archetype predicates in TestDeterminism stay well under this; see spike-a-cel.md.
const CostBudget uint64 = 1000

// NewEnv builds the one activation model used for assert leaves and {{ }} interpolation.
func NewEnv() (*cel.Env, error) {
	return cel.NewEnv(
		ext.NativeTypes(
			reflect.TypeOf(&Facts{}),
			reflect.TypeOf(&QuotaFacts{}),
			reflect.TypeOf(&AuthorFacts{}),
			ext.ParseStructTag("cel"),
		),
		cel.Variable("old", cel.DynType),
		cel.Variable("new", cel.DynType),
		cel.Variable("path", cel.StringType),
		cel.Variable("kind", cel.StringType),
		cel.Variable("entry", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("facts", cel.ObjectType("cel.Facts")),
		cel.Variable("env", cel.StringType),
		// CrossTypeNumericComparisons: int vs double compares at check/eval time.
		// Chosen over adapter-side rewrite of all numbers to double — see report.
		cel.CrossTypeNumericComparisons(true),
		// Standard env only — no time/rand/I/O extensions registered.
	)
}

// ProgramOpts returns program options enforcing the measured cost budget.
func ProgramOpts() []cel.ProgramOption {
	return []cel.ProgramOption{
		cel.CostLimit(CostBudget),
		cel.CostTracking(nil),
	}
}

// CompileExpr compiles a CEL expression against the shared env.
func CompileExpr(env *cel.Env, expr string) (*cel.Ast, error) {
	ast, iss := env.Compile(expr)
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	return ast, nil
}

// EvalBool evaluates a checked AST to a boolean, returning cost when tracked.
func EvalBool(env *cel.Env, ast *cel.Ast, activation any) (bool, uint64, error) {
	prg, err := env.Program(ast, ProgramOpts()...)
	if err != nil {
		return false, 0, err
	}
	out, details, err := prg.Eval(activation)
	var cost uint64
	if details != nil {
		if c := details.ActualCost(); c != nil {
			cost = *c
		}
	}
	if err != nil {
		return false, cost, err
	}
	v, ok := out.Value().(bool)
	if !ok {
		return false, cost, errNotBool
	}
	return v, cost, nil
}
