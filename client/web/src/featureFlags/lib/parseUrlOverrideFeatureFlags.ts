/**
 * Parses and returns feature flag override keys & values
 */
export const parseUrlOverrideFeatureFlags = (queryString: string): Map<string, boolean | null> => {
    const urlParameters = new URLSearchParams(queryString)
    const flags = new Map<string, boolean | null>()

    for (const flag of urlParameters.getAll('feat').flatMap(value => value.split(','))) {
        flags.set(flag.replace(/^(-|~)/, ''), flag.startsWith('~') ? null : !flag.startsWith('-'))
    }

    return flags
}

/**
 * Returns a representation of the feature flags compatible the flag query parameter.
 */
export function formatUrlOverrideFeatureFlags(overrides: Map<string, boolean | null>): string[] {
    const flags: string[] = []
    for (const [flag, value] of overrides) {
        flags.push((value === null ? '~' : value ? '' : '-') + flag)
    }
    return flags
}
