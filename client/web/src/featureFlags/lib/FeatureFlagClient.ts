import { from, iif, Observable, timer } from 'rxjs'
import { distinctUntilChanged, map, retry, shareReplay, switchMap } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import type { requestGraphQL } from '../../backend/graphql'
import { EvaluateFeatureFlagResult } from '../../graphql-operations'
import { FeatureFlagName } from '../featureFlags'

import { getFeatureFlagOverride } from './feature-flag-local-overrides'

// exporting for testing purposes only
const EVALUATE_FEATURE_FLAG_QUERY = gql`
    query EvaluateFeatureFlag($flagName: String!) {
        evaluateFeatureFlag(flagName: $flagName)
    }
`
/**
 * Fetches the evaluated feature flags for the current user
 */
const fetchEvaluateFeatureFlag = (
    requestGraphQLFunc: typeof requestGraphQL,
    flagName: FeatureFlagName
): Observable<EvaluateFeatureFlagResult['evaluateFeatureFlag']> =>
    from(
        requestGraphQLFunc<EvaluateFeatureFlagResult>(EVALUATE_FEATURE_FLAG_QUERY, {
            flagName,
        })
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.evaluateFeatureFlag)
    )

export interface IFeatureFlagClient {
    /**
     * Evaluates and returns feature flag value
     */
    get(flagName: FeatureFlagName): Observable<boolean>
}

export class FeatureFlagClient implements IFeatureFlagClient {
    private cache = new Map<FeatureFlagName, Observable<boolean>>()
    /**
     *
     * @param requestGraphQLFn function to use for making GQL API calls
     * @param interval interval in milliseconds to refetch each feature flag
     */
    constructor(private requestGraphQLFn: typeof requestGraphQL, private interval?: number) {}

    /**
     * For mocking/testing purposes
     *
     * @see {MockedFeatureFlagsProvider}
     */
    public setRequestGraphQLFn(requestGraphQLFn: typeof requestGraphQL): void {
        this.requestGraphQLFn = requestGraphQLFn
    }

    /**
     * Evaluates and returns feature flag once or by interval
     */
    public get(flagName: FeatureFlagName): Observable<boolean> {
        if (!this.cache.has(flagName)) {
            this.cache.set(
                flagName,
                iif(() => typeof this.interval === 'number', timer(0, this.interval), timer(0)).pipe(
                    switchMap(() =>
                        fetchEvaluateFeatureFlag(this.requestGraphQLFn, flagName).pipe(
                            retry(3),
                            map(value => {
                                // Use local feature flag override if exists
                                const overriddenValue = getFeatureFlagOverride(flagName)
                                if (overriddenValue === null) {
                                    return value
                                }
                                if (['true', 1].includes(overriddenValue)) {
                                    return true
                                }
                                if (['false', 0].includes(overriddenValue)) {
                                    return false
                                }
                                return value
                            })
                        )
                    ),
                    distinctUntilChanged(),
                    shareReplay(1)
                )
            )
        }

        return this.cache.get(flagName)!
    }
}
