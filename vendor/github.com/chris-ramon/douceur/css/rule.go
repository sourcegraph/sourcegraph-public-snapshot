package css

import (
	"fmt"
	"strings"
)

const (
	indentSpace = 2
)

// RuleKind represents a Rule kind
type RuleKind int

// Rule kinds
const (
	QualifiedRule RuleKind = iota
	AtRule
)

// At Rules than have Rules inside their block instead of Declarations
var atRulesWithRulesBlock = []string{
	"@document", "@font-feature-values", "@keyframes", "@media", "@supports", "@-webkit-keyframes", "@-moz-keyframes", "@-o-keyframes",
}

// Selector represents a parsed CSS selector.
type Selector struct {
	Value  string
	Line   int
	Column int
}

func (selector Selector) String() string {
	return selector.str(false)
}

func (selector Selector) str(diff bool) string {
	if diff {
		return fmt.Sprintf("Selector: %s (%d, %d)", selector.Value, selector.Line, selector.Column)
	} else {
		return selector.Value
	}
}

// Rule represents a parsed CSS rule
type Rule struct {
	Kind RuleKind

	// At Rule name (eg: "@media")
	Name string

	// Raw prelude
	Prelude string

	// Qualified Rule selectors parsed from prelude
	Selectors []*Selector

	// Style properties
	Declarations []*Declaration

	// At Rule embedded rules
	Rules []*Rule

	// Current rule embedding level
	EmbedLevel int
}

// NewRule instanciates a new Rule
func NewRule(kind RuleKind) *Rule {
	return &Rule{
		Kind: kind,
	}
}

// Returns string representation of rule kind
func (kind RuleKind) String() string {
	switch kind {
	case QualifiedRule:
		return "Qualified Rule"
	case AtRule:
		return "At Rule"
	default:
		return "WAT"
	}
}

// EmbedsRules returns true if this rule embeds another rules
func (rule *Rule) EmbedsRules() bool {
	if rule.Kind == AtRule {
		for _, atRuleName := range atRulesWithRulesBlock {
			if rule.Name == atRuleName {
				return true
			}
		}
	}

	return false
}

// Equal returns true if both rules are equals
func (rule *Rule) Equal(other *Rule) bool {
	if (rule.Kind != other.Kind) ||
		(rule.Prelude != other.Prelude) ||
		(rule.Name != other.Name) {
		return false
	}

	if (len(rule.Selectors) != len(other.Selectors)) ||
		(len(rule.Declarations) != len(other.Declarations)) ||
		(len(rule.Rules) != len(other.Rules)) {
		return false
	}

	for i, sel := range rule.Selectors {
		if (sel.Value != other.Selectors[i].Value) ||
			(sel.Line != other.Selectors[i].Line) ||
			(sel.Column != other.Selectors[i].Column) {
			return false
		}
	}

	for i, decl := range rule.Declarations {
		if !decl.Equal(other.Declarations[i]) {
			return false
		}
	}

	for i, rule := range rule.Rules {
		if !rule.Equal(other.Rules[i]) {
			return false
		}
	}

	return true
}

// Diff returns a string representation of rules differences
func (rule *Rule) Diff(other *Rule) []string {
	result := []string{}

	if rule.Kind != other.Kind {
		result = append(result, fmt.Sprintf("Kind: %s | %s", rule.Kind.String(), other.Kind.String()))
	}

	if rule.Prelude != other.Prelude {
		result = append(result, fmt.Sprintf("Prelude: \"%s\" | \"%s\"", rule.Prelude, other.Prelude))
	}

	if rule.Name != other.Name {
		result = append(result, fmt.Sprintf("Name: \"%s\" | \"%s\"", rule.Name, other.Name))
	}

	if len(rule.Selectors) != len(other.Selectors) {
		result = append(result, fmt.Sprintf("Selectors: %v | %v", strings.Join(rule.Sel(), ", "), strings.Join(other.Sel(), ", ")))
	} else {
		for i, sel := range rule.Selectors {
			if sel != other.Selectors[i] {
				result = append(result, fmt.Sprintf("Selector: \"%s\" | \"%s\"", sel.str(true), other.Selectors[i].str(true)))
			}
		}
	}

	if len(rule.Declarations) != len(other.Declarations) {
		result = append(result, fmt.Sprintf("Declarations Nb: %d | %d", len(rule.Declarations), len(other.Declarations)))
	} else {
		for i, decl := range rule.Declarations {
			if !decl.Equal(other.Declarations[i]) {
				result = append(result, fmt.Sprintf("Declaration: \"%s\" | \"%s\"", decl.Str(true), other.Declarations[i].Str(true)))
			}
		}
	}

	if len(rule.Rules) != len(other.Rules) {
		result = append(result, fmt.Sprintf("Rules Nb: %d | %d", len(rule.Rules), len(other.Rules)))
	} else {

		for i, rule := range rule.Rules {
			if !rule.Equal(other.Rules[i]) {
				result = append(result, fmt.Sprintf("Rule: \"%s\" | \"%s\"", rule.str(true), other.Rules[i].str(true)))
			}
		}
	}

	return result
}

// Returns the string representation of a rule
func (rule *Rule) String() string {
	return rule.str(false)
}

// Returns the string representation of a rule
func (rule *Rule) str(diff bool) string {
	result := ""

	if rule.Kind == QualifiedRule {
		for i, sel := range rule.Sel() {
			if i != 0 {
				result += ", "
			}
			result += sel
		}
	} else {
		// AtRule
		result += fmt.Sprintf("%s", rule.Name)

		if rule.Prelude != "" {
			if result != "" {
				result += " "
			}
			result += fmt.Sprintf("%s", rule.Prelude)
		}
	}

	if (len(rule.Declarations) == 0) && (len(rule.Rules) == 0) {
		result += ";"
	} else {
		result += " {\n"

		if rule.EmbedsRules() {
			for _, subRule := range rule.Rules {
				result += fmt.Sprintf("%s%s\n", rule.indent(), subRule.str(diff))
			}
		} else {
			for _, decl := range rule.Declarations {
				result += fmt.Sprintf("%s%s\n", rule.indent(), decl.Str(diff))
			}
		}

		result += fmt.Sprintf("%s}", rule.indentEndBlock())
	}

	return result
}

// Returns identation spaces for declarations and rules
func (rule *Rule) indent() string {
	return strings.Repeat(" ", (rule.EmbedLevel+1)*indentSpace)
}

// Returns identation spaces for end of block character
func (rule *Rule) indentEndBlock() string {
	return strings.Repeat(" ", rule.EmbedLevel*indentSpace)
}

func (rule *Rule) Sel() []string {
	selectors := []string{}
	for _, s := range rule.Selectors {
		selectors = append(selectors, s.Value)
	}
	return selectors
}
