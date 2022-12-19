import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import type { requestGraphQL } from '../../backend/graphql'
import { EvaluateFeatureFlagResult } from '../../graphql-operations'
import { FeatureFlagName } from '../featureFlags'

import { getFeatureFlagOverrideValue } from './feature-flag-local-overrides'

/**
 * Evaluate feature flags for the current user
 */
const fetchEvaluateFeatureFlag = (
    requestGraphQLFunc: typeof requestGraphQL,
    flagName: FeatureFlagName
): Promise<EvaluateFeatureFlagResult['evaluateFeatureFlag']> =>
    requestGraphQLFunc<EvaluateFeatureFlagResult>(
        gql`
            query EvaluateFeatureFlag($flagName: String!) {
                evaluateFeatureFlag(flagName: $flagName)
            }
        `,
        {
            flagName,
        }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.evaluateFeatureFlag)
        )
        .toPromise()

/**
 * Feature flag client service. Should be used as singleton for the whole application.
 */
export class FeatureFlagClient {
    private flags = new Map<FeatureFlagName, Promise<EvaluateFeatureFlagResult['evaluateFeatureFlag']>>()

    /**
     * @param requestGraphQLFunction function to use for making GQL API calls.
     * @param cacheTimeToLive milliseconds to keep the value in the in-memory client-side cache.
     */
    constructor(private requestGraphQLFunction: typeof requestGraphQL, private cacheTimeToLive?: number) {}

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
    public get(flagName: FeatureFlagName): Promise<EvaluateFeatureFlagResult['evaluateFeatureFlag']> {
        if (!this.flags.has(flagName)) {
            const overriddenValue = getFeatureFlagOverrideValue(flagName)

            // Use local feature flag override if exists
            if (overriddenValue !== null) {
                return Promise.resolve(overriddenValue)
            }

            const flag = fetchEvaluateFeatureFlag(this.requestGraphQLFunction, flagName)
            this.flags.set(flagName, flag)

            if (this.cacheTimeToLive) {
                setTimeout(() => this.flags.delete(flagName), this.cacheTimeToLive)
            }
        }

        return this.flags.get(flagName)!
    }
}
