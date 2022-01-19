/**
 * Parses and returns feature flag override keys & values
 */
export const parseUrlOverrideFeatureFlags = (queryString: string): Record<string, string | undefined> | undefined => {
    const urlParameters = new URLSearchParams(queryString)
    urlParameters.get('feature-flag-key') // ?
    const overrideFeatureFlagKeys = (urlParameters.get('feature-flag-key') ?? '').split(/\s*,\s*/g).filter(Boolean) // ?
    if (overrideFeatureFlagKeys.length === 0) {
        return
    }
    const overrideFeatureFlagValues = (urlParameters.get('feature-flag-value') ?? '').split(/\s*,\s*/g).filter(Boolean)
    return overrideFeatureFlagKeys.reduce(
        (result, key, index) => ({ ...result, [key]: overrideFeatureFlagValues[index] }),
        {}
    )
}

