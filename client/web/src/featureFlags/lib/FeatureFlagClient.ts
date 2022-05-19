import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../backend/graphql'
import { FetchFeatureFlagsResult } from '../../graphql-operations'
import { FeatureFlagName } from '../featureFlags'

import { getFeatureFlagOverride } from './feature-flag-local-overrides'

/**
 * Fetches the evaluated feature flags for the current user
 */
function fetchFeatureFlags(): Observable<FetchFeatureFlagsResult['viewerFeatureFlags']> {
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
    on(flagName: FeatureFlagName, callback: (value: boolean, error?: Error) => void): () => void
}

export class FeatureFlagClient implements IFeatureFlagClient {
    private cache = new FeatureFlagsProxyMap()
    constructor() {
        fetchFeatureFlags()
            .toPromise()
            .then(flags => {
                for (const flag of flags) {
                    this.cache.set(flag.name as FeatureFlagName, flag.value)
                }
            })
            .catch(console.error)
    }

    /**
     * Calls callback function whenever a feature flag is changed
     * NOTE: Will change to actual per flag evaluation and listeners on flag value change once backend is migrated to new API
     *
     * @returns a cleanup/unsubscribe function
     */
    // eslint-disable-next-line id-length
    public on(flagName: FeatureFlagName, callback: (value: boolean, error?: Error) => void): () => void {
        Promise.resolve(this.cache.get(flagName) || false)
            .then(callback)
            .catch(error => callback(false, error))
        return () => {
            // TODO: cleanup
        }
    }
}
