//go:build !no_runtime_type_checking

package constructs

import (
	"fmt"
)

func validateDependable_GetParameters(instance IDependable) error {
	if instance == nil {
		return fmt.Errorf("parameter instance is required, but nil was provided")
	}

	return nil
}

func validateDependable_ImplementParameters(instance IDependable, trait Dependable) error {
	if instance == nil {
		return fmt.Errorf("parameter instance is required, but nil was provided")
	}

	if trait == nil {
		return fmt.Errorf("parameter trait is required, but nil was provided")
	}

	return nil
}

func validateDependable_OfParameters(instance IDependable) error {
	if instance == nil {
		return fmt.Errorf("parameter instance is required, but nil was provided")
	}

	return nil
}

