package symbolexample

import (
	"github.com/sourcegraph/sourcegraph/schema"
)

func AuthProviderType(p schema.AuthProviders) string {
	switch {
	case p.Builtin != nil:
		return p.Builtin.Type
	case p.Gitlab != nil:
		return p.Gitlab.Type
	default:
		return ""
	}
}

func authAllowSignup(c *Unified) bool {
	for _, p := range c.AuthProviders {
		if p.Builtin != nil && p.Builtin.AllowSignup {
			return true
		}
	}
	return false
}

func MadeUp() SomeSymbol {}

func CallsAFunction() bool {
	x := DoSomething()
	y := pkg.DoSomething()
}
