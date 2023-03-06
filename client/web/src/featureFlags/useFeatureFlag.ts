import { logger } from '@sourcegraph/common'
import { getDocumentNode, gql, useQuery } from '@sourcegraph/http-client'

import { EvaluateFeatureFlagResult, EvaluateFeatureFlagVariables } from '../graphql-operations'

import { FeatureFlagName } from './featureFlags'
import { getFeatureFlagOverrideValue } from './lib/feature-flag-local-overrides'

export const EVALUATE_FEATURE_FLAG_QUERY = getDocumentNode(gql`
    query EvaluateFeatureFlag($flagName: String!) {
        evaluateFeatureFlag(flagName: $flagName)
    }
`)

type FetchStatus = 'initial' | 'loaded' | 'error'
type FlagName = string
type EvaluateError = unknown

const MINUTE = 60000
const FEATURE_FLAG_CACHE_TTL = MINUTE * 10

/**
 * 1. [undefined] We don't have flag value yet -> its name is not present in this Map.
 *
 * 2. [valid] We got the fresh value of the flag from the API -> name is added to this map
 * with the `valid` status. It means that there's an in-flight `setTimeout` for
 * this flag which will make it `stale` after `cacheTTL` ms.
 *
 * 3. [stale] The flag TTL just elapsed -> the flag state is set to `stale`, and refetch is called.
 *
 * 4. [refetch] The refetch is completed -> the flag state is removed to restart the cycle.
 */
const FLAG_STATES = new Map<FlagName, 'valid' | 'stale' | 'refetch'>()

/**
 * Evaluates the feature flag via GraphQL query and returns the value.
 * Prioritizes feature flag overrides. It uses the value cached by Apollo
 * Client if available and not stale. The hook render starts the timeout,
 * which ensures that we mark the cached value as stale after `cacheTTL`.
 * If the cache TTL is elapsed on the next render, we refetch the value.
 */
export function useFeatureFlag(
    flagName: FeatureFlagName,
    defaultValue = false,
    cacheTTL = FEATURE_FLAG_CACHE_TTL,
    flagStates = FLAG_STATES
): [boolean, FetchStatus, EvaluateError?] {
    const overriddenValue = getFeatureFlagOverrideValue(flagName)
    const overriddenValueExists = typeof overriddenValue === 'boolean'

    const { data, error, refetch } = useQuery<EvaluateFeatureFlagResult, EvaluateFeatureFlagVariables>(
        EVALUATE_FEATURE_FLAG_QUERY,
        {
            variables: { flagName },
            fetchPolicy: 'cache-first',
            nextFetchPolicy: 'network-only',
            skip: overriddenValueExists,
        }
    )

    /**
     * Skip the GraphQL query and return the `overriddenValue` if it exists.
     */
    if (overriddenValueExists) {
        return [overriddenValue, 'loaded']
    }

    if (data) {
        switch (flagStates.get(flagName)) {
            /**
             * Do nothing. Either `setTimeout` or `refetch` is in progress,
             * and we can use the cached value for now.
             */
            case 'valid':
            case 'refetch':
                break

            /**
             * If we have the stale value, initiate the refetch with the `network-only`
             * strategy. The hook will be re-rendered only if the value has changed after the refetch.
             * On refetch success, we mark the value as the new one by removing the flag status.
             */
            case 'stale': {
                flagStates.set(flagName, 'refetch')
                refetch()
                    .then(() => flagStates.delete(flagName))
                    .catch(logger.error)
                break
            }

            /**
             * After receiving the new value from the API, we start the timer unique
             * for the feature flag, marking it as stale after `cacheTTL`.
             */
            default: {
                flagStates.set(flagName, 'valid')
                setTimeout(() => flagStates.set(flagName, 'stale'), cacheTTL)
            }
        }
    }

    const value = data?.evaluateFeatureFlag ?? defaultValue
    const status = error ? 'error' : data ? 'loaded' : 'initial'

    return [value, status, error?.networkError]
}
