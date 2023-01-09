const buildFlagOverrideKey = (key: string): string => `featureFlagOverride-${key}`

export const getFeatureFlagOverride = (flagName: string): string | null =>
    localStorage.getItem(buildFlagOverrideKey(flagName))

export const setFeatureFlagOverride = (flagName: string, value: boolean): void =>
    localStorage.setItem(buildFlagOverrideKey(flagName), value.toString())

export const removeFeatureFlagOverride = (flagName: string): void =>
    localStorage.removeItem(buildFlagOverrideKey(flagName))

export const getFeatureFlagOverrideValue = (flagName: string): boolean | null => {
    const overriddenValue = getFeatureFlagOverride(flagName)

    if (overriddenValue === null) {
        return null
    }

    if (['true', 1].includes(overriddenValue)) {
        return true
    }

    if (['false', 0].includes(overriddenValue)) {
        return false
    }

    return null
}
