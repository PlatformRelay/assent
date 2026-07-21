package celspike

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/cel-go/cel"
)

var interpSlot = regexp.MustCompile(`\{\{\s*([^}]+?)\s*\}\}`)

// InterpSlot is one {{ expr }} region in a message template.
type InterpSlot struct {
	Expr   string
	Offset int // byte offset of '{{' in the template
	Line   int
	Column int
}

// ParseMessageTemplate extracts CEL slots from a message string.
func ParseMessageTemplate(tmpl string) []InterpSlot {
	matches := interpSlot.FindAllStringSubmatchIndex(tmpl, -1)
	out := make([]InterpSlot, 0, len(matches))
	for _, m := range matches {
		expr := tmpl[m[2]:m[3]]
		offset := m[0]
		line, col := position(tmpl, offset)
		out = append(out, InterpSlot{
			Expr:   strings.TrimSpace(expr),
			Offset: offset,
			Line:   line,
			Column: col,
		})
	}
	return out
}

func position(s string, offset int) (line, col int) {
	line, col = 1, 1
	for i := 0; i < offset && i < len(s); i++ {
		if s[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

// CompileMessage checks every {{ }} slot against the shared activation env.
// Unknown fields fail at compile time with a positioned error — never "<no value>".
func CompileMessage(env *cel.Env, tmpl string) error {
	slots := ParseMessageTemplate(tmpl)
	for _, slot := range slots {
		_, iss := env.Compile(slot.Expr)
		if iss.Err() != nil {
			// Prefer cel-go's own location when present; fall back to template offset.
			for _, e := range iss.Errors() {
				line, col := slot.Line, slot.Column
				if e.Location != nil {
					// Location is relative to the slot expression; shift to template.
					if e.Location.Line() == 1 {
						col = slot.Column + e.Location.Column()
					} else {
						line = slot.Line + e.Location.Line() - 1
						col = e.Location.Column()
					}
				}
				return fmt.Errorf("message template:%d:%d: %s", line, col, e.Message)
			}
			return fmt.Errorf("message template:%d:%d: %w", slot.Line, slot.Column, iss.Err())
		}
	}
	return nil
}
