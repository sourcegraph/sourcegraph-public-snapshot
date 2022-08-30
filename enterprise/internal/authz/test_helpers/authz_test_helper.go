const FeatureACLSErr = "The feature \"acls\" is not activated because it requires a valid Sourcegraph license. Purchase a Sourcegraph subscription to activate this feature."

func MockLicenseCheckErr(expectedError string) {
	licensing.MockCheckFeature = func(feature licensing.Feature) error {
		if expectedError == "" {
			return nil
		}
		return errors.New(expectedError)
	}
}
