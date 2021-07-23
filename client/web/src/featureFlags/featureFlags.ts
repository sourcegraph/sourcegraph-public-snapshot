import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../backend/graphql'
import { FetchFeatureFlagsResult } from '../graphql-operations'

// This should be a `|` separated union of feature flag strings,
// but right now we don't have any feature flags being tested.
export type FeatureFlagName = never

export type FlagSet = Map<FeatureFlagName, boolean>

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
            const result = new Map<FeatureFlagName, boolean>()
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
