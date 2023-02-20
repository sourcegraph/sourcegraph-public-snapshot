import { dataOrThrowErrors, getDocumentNode, gql, GraphQLClient } from '@sourcegraph/http-client'

import { EvaluateFeatureFlagResult, EvaluateFeatureFlagVariables } from '../../graphql-operations'
import { FeatureFlagName } from '../featureFlags'

import { getFeatureFlagOverrideValue } from './feature-flag-local-overrides'

/**
 * TODO: do we want to use the default `useQuery` and `gql` exports from `@apollo/client`
 * instead of this legacy setup?
 */
const query = getDocumentNode(gql`
    query EvaluateFeatureFlag($flagName: String!) {
        evaluateFeatureFlag(flagName: $flagName) {
            id
            name
            value
        }
    }
`)

/**
 * Feature flag client service. Should be used as singleton for the whole application.
 */
export class FeatureFlagClient {
    private flags = new Map<string, boolean>()
    private graphqlClient?: GraphQLClient

    /**
     * @param requestGraphQLFunction function to use for making GQL API calls.
     * @param cacheTimeToLive milliseconds to keep the value in the in-memory client-side cache.
     */
    constructor(getGraphqlClient: () => Promise<GraphQLClient>, private cacheTimeToLive?: number) {
        getGraphqlClient().then(client => {
            this.graphqlClient = client
        })
    }

    /**
     * For mocking/testing purposes
     *
     * @see {MockedFeatureFlagsProvider}
     */
    // public setRequestGraphQLFunction(requestGraphQLFunction: GraphQLClient['query']): void {
    //     this.getGraphqlClient = () => Promise.resolve({ query: requestGraphQLFunction })
    // }

    private scheduleCacheEvict(flag: EvaluateFeatureFlagResult['evaluateFeatureFlag']) {
        const { id, name } = flag

        if (this.cacheTimeToLive && !this.flags.get(name)) {
            this.flags.set(name, true)

            setTimeout(() => {
                this.graphqlClient?.cache.evict({ id })
            }, this.cacheTimeToLive)
        }
    }

    /**
     * Evaluates and returns feature flag value
     */
    public get(flagName: FeatureFlagName): Promise<EvaluateFeatureFlagResult['evaluateFeatureFlag']['value']> {
        if (!this.graphqlClient) {
            throw new Error('oops')
        }

        const overriddenValue = getFeatureFlagOverrideValue(flagName)

        // Use local feature flag override if exists
        if (overriddenValue !== null) {
            return Promise.resolve(overriddenValue)
        }

        const cachedResult = this.graphqlClient.readQuery<EvaluateFeatureFlagResult, EvaluateFeatureFlagVariables>({
            query,
            variables: {
                flagName,
            },
        })

        if (cachedResult) {
            this.scheduleCacheEvict(cachedResult.evaluateFeatureFlag)

            return Promise.resolve(cachedResult.evaluateFeatureFlag.value)
        }

        return this.graphqlClient
            .query<EvaluateFeatureFlagResult, EvaluateFeatureFlagVariables>({
                query,
                variables: {
                    flagName,
                },
            })
            .then(dataOrThrowErrors)
            .then(data => {
                this.scheduleCacheEvict(data.evaluateFeatureFlag)
                return data.evaluateFeatureFlag.value
            })
    }
}
