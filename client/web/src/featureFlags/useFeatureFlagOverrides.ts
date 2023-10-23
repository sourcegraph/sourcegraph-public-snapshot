import { useMemo } from 'react'

import { useOverrideCounter } from '../stores'

import { getFeatureFlagOverrides } from './lib/feature-flag-local-overrides'

/**
 * Hook for subscribing to all current feature flag overrides.
 */
export function useFeatureFlagOverrides(): Map<string, boolean> {
    // Local feature flag overrides are stored in local storage.
    // Because there is no direct way to be informed about when
    // these values change (without a larger refactor of how
    // these work), we use the override counter as a workaround.
    // But this only works because at is is used now, the counter
    // value gets updated every time the value of an override changes.

    const counter = useOverrideCounter()
    const overrides = useMemo(
        () =>
            // `counter` is referenced here to make linters happy
            // Better than disabling hook dependency checks IMO.
            counter.featureFlags ? getFeatureFlagOverrides() : new Map(),
        [counter]
    )

    return overrides
}
