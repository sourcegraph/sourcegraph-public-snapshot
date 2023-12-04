import { gql, query } from '$lib/graphql'
import type { EvaluatedFeatureFlagsResult } from '$lib/graphql-operations'

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
export async function fetchEvaluatedFeatureFlags(): Promise<FeatureFlag[]> {
    return (
        await query<EvaluatedFeatureFlagsResult>(FEATUREFLAGS_QUERY, undefined, {
            fetchPolicy: 'no-cache',
        })
    ).evaluatedFeatureFlags
}
