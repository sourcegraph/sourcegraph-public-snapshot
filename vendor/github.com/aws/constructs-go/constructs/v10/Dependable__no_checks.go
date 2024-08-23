//go:build no_runtime_type_checking

package constructs

// Building without runtime type checking enabled, so all the below just return nil

func validateDependable_GetParameters(instance IDependable) error {
	return nil
}

func validateDependable_ImplementParameters(instance IDependable, trait Dependable) error {
	return nil
}

func validateDependable_OfParameters(instance IDependable) error {
	return nil
}

