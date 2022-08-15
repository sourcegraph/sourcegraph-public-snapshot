import { iif, Observable, timer } from 'rxjs'
import { distinctUntilChanged, map, retry, shareReplay, switchMap } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import type { requestGraphQL } from '../../backend/graphql'
import { EvaluateFeatureFlagResult } from '../../graphql-operations'
import { FeatureFlagName } from '../featureFlags'

import { getFeatureFlagOverride } from './feature-flag-local-overrides'

/**
 * Evaluate feature flags for the current user
 */
const fetchEvaluateFeatureFlag = (
    requestGraphQLFunc: typeof requestGraphQL,
    flagName: FeatureFlagName
): Observable<EvaluateFeatureFlagResult['evaluateFeatureFlag']> =>
    requestGraphQLFunc<EvaluateFeatureFlagResult>(
        gql`
            query EvaluateFeatureFlag($flagName: String!) {
                evaluateFeatureFlag(flagName: $flagName)
            }
        `,
        {
            flagName,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.evaluateFeatureFlag)
    )

/**
 * Feature flag client service. Should be used as singleton for the whole application.
 */
export class FeatureFlagClient {
    private flags = new Map<FeatureFlagName, Observable<EvaluateFeatureFlagResult['evaluateFeatureFlag']>>()

    /**
     * @param requestGraphQLFunction function to use for making GQL API calls
     * @param refetchInterval milliseconds to refetch each feature flag evaluation. Fetches once if undefined provided.
     */
    constructor(private requestGraphQLFunction: typeof requestGraphQL, private refetchInterval?: number) {}

    /**
     * For mocking/testing purposes
     *
     * @see {MockedFeatureFlagsProvider}
     */
    public setRequestGraphQLFunction(requestGraphQLFunction: typeof requestGraphQL): void {
        this.requestGraphQLFunction = requestGraphQLFunction
    }

    /**
     * Evaluates and returns feature flag value
     */
    public get(flagName: FeatureFlagName): Observable<EvaluateFeatureFlagResult['evaluateFeatureFlag']> {
        if (!this.flags.has(flagName)) {
            const flag$ = iif(
                () => typeof this.refetchInterval === 'number' && this.refetchInterval > 0,
                timer(0, this.refetchInterval),
                timer(0)
            ).pipe(
                switchMap(() => fetchEvaluateFeatureFlag(this.requestGraphQLFunction, flagName).pipe(retry(3))),
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
                }),
                distinctUntilChanged(),
                // shared between all subscribers to avoid multiple computations/API calls
                shareReplay(1)
            )

            this.flags.set(flagName, flag$)
        }

        return this.flags.get(flagName)!
    }
}
