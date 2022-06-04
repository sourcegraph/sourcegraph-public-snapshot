import { from, Observable, ReplaySubject } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import type { requestGraphQL } from '../../backend/graphql'
import { FetchFeatureFlagsResult } from '../../graphql-operations'
import { FeatureFlagName } from '../featureFlags'

import { getFeatureFlagOverride } from './feature-flag-local-overrides'

/**
 * Fetches the evaluated feature flags for the current user
 */
function fetchFeatureFlags(
    requestGraphQLFunction: typeof requestGraphQL
): Observable<FetchFeatureFlagsResult['viewerFeatureFlags']> {
    return from(
        requestGraphQLFunction<FetchFeatureFlagsResult>(
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
        map(data => data.viewerFeatureFlags)
    )
}

/**
 * A Map wrapper that first looks in localStorage.get when returning values.
 */
class FeatureFlagsProxyMap extends Map<FeatureFlagName, boolean> {
    public get(key: FeatureFlagName): boolean | undefined {
        const overriddenValue = getFeatureFlagOverride(key)

        if (overriddenValue !== null && ['true', 'false'].includes(overriddenValue)) {
            return JSON.parse(overriddenValue) as boolean
        }

        return super.get(key)
    }
}

export interface IFeatureFlagClient {
    get(flagName: FeatureFlagName): Observable<boolean>
}

export class FeatureFlagClient implements IFeatureFlagClient {
    private flags$ = new ReplaySubject<FeatureFlagsProxyMap>(1)
    constructor(requestGraphQLFunction: typeof requestGraphQL) {
        fetchFeatureFlags(requestGraphQLFunction)
            .toPromise()
            .then(flags =>
                this.flags$.next(
                    new FeatureFlagsProxyMap(flags.map(({ name, value }) => [name as FeatureFlagName, value]))
                )
            )
            .catch(error => {
                console.error(error)
                this.flags$.error(error)
            })
    }

    /**
     * Calls callback function whenever a feature flag is changed
     * NOTE: Will change to actual per flag evaluation and listeners on flag value change once backend is migrated to new API
     *
     * @returns a cleanup/unsubscribe function
     */
    public get(flagName: FeatureFlagName): Observable<boolean> {
        return this.flags$.pipe(
            map(flags => flags.get(flagName) || false),
            distinctUntilChanged()
        )
    }
}
