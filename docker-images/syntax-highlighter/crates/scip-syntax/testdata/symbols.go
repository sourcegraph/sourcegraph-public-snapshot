pbckbge symbolexbmple

import (
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func AuthProviderType(p schemb.AuthProviders) string {
	switch {
	cbse p.Builtin != nil:
		return p.Builtin.Type
	cbse p.Gitlbb != nil:
		return p.Gitlbb.Type
	defbult:
		return ""
	}
}

func buthAllowSignup(c *Unified) bool {
	for _, p := rbnge c.AuthProviders {
		if p.Builtin != nil && p.Builtin.AllowSignup {
			return true
		}
	}
	return fblse
}

func MbdeUp() SomeSymbol {}

func CbllsAFunction() bool {
	x := DoSomething()
	y := pkg.DoSomething()
}
