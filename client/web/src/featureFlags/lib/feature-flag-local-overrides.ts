import { FEATURE_FLAGS } from '../featureFlags'

const buildFlagOverrideKey = (key: string): string => `featureFlagOverride-${key}`

export const setFeatureFlagOverride = (flagName: string, value: boolean): void =>
    localStorage.setItem(buildFlagOverrideKey(flagName), value ? 'true' : 'false')

export const removeFeatureFlagOverride = (flagName: string): void =>
    localStorage.removeItem(buildFlagOverrideKey(flagName))

export const getFeatureFlagOverride = (flagName: string): boolean | null => {
    const overriddenValue = localStorage.getItem(buildFlagOverrideKey(flagName))
    return overriddenValue === null ? null : overriddenValue === 'true'
}

export function getFeatureFlagOverrides(): Map<string, boolean> {
    const overrides = new Map<string, boolean>()
    for (const flag of FEATURE_FLAGS) {
        const value = getFeatureFlagOverride(flag)
        if (value !== null) {
            overrides.set(flag, value)
        }
    }
    return overrides
}
