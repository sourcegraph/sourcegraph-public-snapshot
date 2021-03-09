package batches

func featuresAllEnabled() featureFlags {
	return featureFlags{
		allowArrayEnvironments:   true,
		includeAutoAuthorDetails: true,
		useGzipCompression:       true,
	}
}
