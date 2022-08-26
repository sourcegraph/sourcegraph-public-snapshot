import * as sourcegraph from '@sourcegraph/extension-api-types'

import { SearchBasedCodeIntelligenceSettings } from './settings'

/** Retrieves a config value by key. */
export function getConfig<K extends keyof SearchBasedCodeIntelligenceSettings>(
    key: K,
    defaultValue: NonNullable<SearchBasedCodeIntelligenceSettings[K]>
): NonNullable<SearchBasedCodeIntelligenceSettings[K]> {
    const value = sourcegraph.configuration.get<SearchBasedCodeIntelligenceSettings>().get(key)
    if (value === undefined) {
        return defaultValue
    }
    return value as NonNullable<SearchBasedCodeIntelligenceSettings[K]>
}
