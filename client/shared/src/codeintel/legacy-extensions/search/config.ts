import * as sourcegraph from '../api'

import type { SearchBasedCodeIntelligenceSettings } from './settings'

/** Retrieves a config value by key. */
export function getConfig<K extends keyof SearchBasedCodeIntelligenceSettings>(
    key: K,
    defaultValue: NonNullable<SearchBasedCodeIntelligenceSettings[K]>
): NonNullable<SearchBasedCodeIntelligenceSettings[K]> {
    const value = sourcegraph.getSetting<SearchBasedCodeIntelligenceSettings>(key)
    if (value === undefined) {
        return defaultValue
    }
    return value as NonNullable<SearchBasedCodeIntelligenceSettings[K]>
}
