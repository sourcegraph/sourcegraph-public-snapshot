import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../backend/graphql'
import { FetchFeatureFlagsResult } from '../graphql-operations'

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
            if (!data) {
                return {}
            }
            const result: FlagSet = {}
            for (const flag of data.viewerFeatureFlags) {
                result[flag.name] = flag.value
            }
            return result
        })
    )
}

export interface FlagSet {
    [key: string]: boolean
}
