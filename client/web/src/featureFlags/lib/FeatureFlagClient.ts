import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../backend/graphql'
import { EvaluatedFeatureFlagsResult, EvaluateFeatureFlagResult } from '../../graphql-operations'
import { FeatureFlagName } from '../featureFlags'

import { getFeatureFlagOverride } from './feature-flag-local-overrides'

/**
 * Fetches the evaluated feature flags for the current user
 */
const fetchEvaluatedFeatureFlags = (): Observable<EvaluatedFeatureFlagsResult['evaluatedFeatureFlags']> =>
    from(
        requestGraphQL<EvaluatedFeatureFlagsResult>(
            gql`
                query EvaluatedFeatureFlags {
                    evaluatedFeatureFlags {
                        name
                        value
                    }
                }
            `
        )
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.evaluatedFeatureFlags)
    )

/**
 * Fetches the evaluated feature flags for the current user
 */
const fetchEvaluateFeatureFlag = (
    flagName: FeatureFlagName
): Observable<EvaluateFeatureFlagResult['evaluateFeatureFlag']> =>
    from(
        requestGraphQL<EvaluateFeatureFlagResult>(
            gql`
                query EvaluateFeatureFlag($flagName: String!) {
                    evaluateFeatureFlag(flagName: $flagName)
                }
            `,
            {
                flagName,
            }
        )
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.evaluateFeatureFlag)
    )

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

type FeatureFlagListener = (value: boolean, error?: Error) => void

export class FeatureFlagClient implements IFeatureFlagClient {
    private cache = new FeatureFlagsProxyMap()
    private listeners = new Map<FeatureFlagName, Set<FeatureFlagListener>>()
    constructor() {
        fetchEvaluatedFeatureFlags()
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
     *
     * @returns a cleanup/unsubscribe function
     */
    // eslint-disable-next-line id-length
    public on(flagName: FeatureFlagName, callback: FeatureFlagListener): () => void {
        if (!this.listeners.has(flagName)) {
            this.listeners.set(flagName, new Set())
        }
        this.listeners.get(flagName)?.add(callback)

        fetchEvaluateFeatureFlag(flagName)
            .toPromise()
            .then(flagValue => this.cache.set(flagName, flagValue))
            .catch(error => callback(this.cache.get(flagName) || false, error))
            .finally(() => {
                if (!this.listeners.has(flagName)) {
                    return
                }
                for (const listener of this.listeners.get(flagName)!) {
                    listener(this.cache.get(flagName) || false)
                }
            })
        return () => this.listeners.get(flagName)?.delete(callback)
    }
}
