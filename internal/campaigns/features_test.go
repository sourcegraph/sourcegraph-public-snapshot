package campaigns

func featuresAllEnabled() featureFlags {
	return featureFlags{
		includeAutoAuthorDetails: true,
		useGzipCompression:       true,
	}
}
