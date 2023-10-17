import { derived, readable, type Readable } from 'svelte/store'

import { getStores } from '$lib/stores'
import type { FeatureFlagName } from '$lib/web'

import type { FeatureFlag } from './api'

const MINUTE = 60000
const FEATURE_FLAG_CACHE_TTL = MINUTE * 10

const defaultValues: Partial<Record<FeatureFlagName, boolean>> = {
    'repository-metadata': true,
}

export function createFeatureFlagStore(
    initialFeatureFlags: FeatureFlag[],
    fetchEvaluatedFeatureFlags: () => Promise<FeatureFlag[]>
): Readable<FeatureFlag[]> {
    return readable<FeatureFlag[]>(initialFeatureFlags, set => {
        const timer = globalThis.setInterval(() => {
            fetchEvaluatedFeatureFlags().then(set)
        }, FEATURE_FLAG_CACHE_TTL)

        return () => {
            globalThis.clearInterval(timer)
        }
    })
}

export function featureFlag(name: FeatureFlagName): Readable<boolean> {
    // TODO: add support for overrides
    return derived(
        getStores().featureFlags,
        $featureFlags => $featureFlags.find(flag => flag.name === name)?.value ?? defaultValues[name] ?? false
    )
}
