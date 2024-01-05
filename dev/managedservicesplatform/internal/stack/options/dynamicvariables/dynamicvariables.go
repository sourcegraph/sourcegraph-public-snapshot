package dynamicvariables

import (
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
)

// With generates TFVars using makeVars if we are in a non-stable generation
// mode, and assigns it to the stack.
//
// Values can be referenced using resource/tfvar.
func With(stable bool, makeVars func() (stack.TFVars, error)) stack.NewStackOption {
	return func(s stack.Stack) error {
		if stable {
			return nil
		}
		vars, err := makeVars()
		if err != nil {
			return err
		}
		for k, v := range vars {
			s.DynamicVariables[k] = v
		}
		return nil
	}
}
