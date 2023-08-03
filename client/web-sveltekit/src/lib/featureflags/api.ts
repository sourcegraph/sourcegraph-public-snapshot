import type { EvaluatedFeatureFlagsResult } from '$lib/graphql-operations'
import { dataOrThrowErrors, getDocumentNode, gql, type GraphQLClient } from '$lib/http-client'

export interface FeatureFlag {
    name: string
    value: boolean
}

const FEATUREFLAGS_QUERY = gql`
    query EvaluatedFeatureFlags {
        evaluatedFeatureFlags {
            name
            value
        }
    }
`
export async function fetchEvaluatedFeatureFlags(client: GraphQLClient): Promise<FeatureFlag[]> {
    return dataOrThrowErrors(
        await client.query<EvaluatedFeatureFlagsResult>({
            query: getDocumentNode(FEATUREFLAGS_QUERY),
        })
    ).evaluatedFeatureFlags
}
