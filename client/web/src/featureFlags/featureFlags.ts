import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../backend/graphql'
import { FetchFeatureFlagsResult } from '../graphql-operations'

import { getOverrideKey } from './lib/getOverrideKey'

class ProxyMap<K extends string, V extends boolean> extends Map<K, V> {
    constructor(private getter?: (key: K, value: V | undefined) => V | undefined) {
        super()
    }
    public get(key: K): V | undefined {
        const originalValue = super.get(key)
        return this.getter ? this.getter(key, originalValue) : originalValue
    }
}

// A union of all feature flags we currently have.
// If there are no feature flags at the moment, this should be `never`.
export type FeatureFlagName =
    | 'search-notebook-onboarding'
    | 'test-flag'
    | 'signup-optimization'
    | 'getting-started-tour'

export type FlagSet = ProxyMap<FeatureFlagName, boolean>

/**
 * Fetches the evaluated feature flags for the current user
 */
export function fetchFeatureFlags(): Observable<FlagSet> {
    return from(
        requestGraphQL<FetchFeatureFlagsResult>(
            gql`
                query FetchFeatureFlags {
                    viewerFeatureFlags {
                        name
                        value
                    }
                }
            `
        )
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            const result = new ProxyMap<FeatureFlagName, boolean>((key: FeatureFlagName, value?: boolean):
                | boolean
                | undefined => {
                const overriddenValue = localStorage.getItem(getOverrideKey(key))

                return overriddenValue !== 'undefined' &&
                    overriddenValue !== null &&
                    ['true', 'false'].includes(overriddenValue)
                    ? (JSON.parse(overriddenValue) as boolean)
                    : value
            })
            for (const flag of data.viewerFeatureFlags) {
                result.set(flag.name as FeatureFlagName, flag.value)
            }
            return result
        })
    )
}

export interface FeatureFlagProps {
    /**
     * Evaluated feature flags for the current viewer
     */
    featureFlags: FlagSet
}

export const EMPTY_FEATURE_FLAGS = new Map<FeatureFlagName, boolean>()
