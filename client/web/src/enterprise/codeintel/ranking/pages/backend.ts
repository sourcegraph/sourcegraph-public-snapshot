import type { ApolloError, ApolloQueryResult } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'

import type { RankingSummaryResult, RankingSummaryVariables } from '../../../../graphql-operations'

export const RankingSummaryFieldsFragment = gql`
    fragment RankingSummaryFields on RankingSummary {
        graphKey
        visibleToZoekt
        pathMapperProgress {
            ...RankingSummaryProgressFields
        }
        referenceMapperProgress {
            ...RankingSummaryProgressFields
        }
        reducerProgress {
            ...RankingSummaryProgressFields
        }
    }

    fragment RankingSummaryProgressFields on RankingSummaryProgress {
        startedAt
        completedAt
        processed
        total
    }
`

export const RANKING_SUMMARY = gql`
    query RankingSummary {
        rankingSummary {
            rankingSummary {
                ...RankingSummaryFields
            }
            derivativeGraphKey
            nextJobStartsAt
            numExportedIndexes
            numTargetIndexes
            numRepositoriesWithoutCurrentRanks
        }
    }

    ${RankingSummaryFieldsFragment}
`

export const useRankingSummary = (
    variables: RankingSummaryVariables
): {
    error?: ApolloError
    loading: boolean
    data: RankingSummaryResult | undefined
    refetch: () => Promise<ApolloQueryResult<RankingSummaryResult>>
} =>
    useQuery<RankingSummaryResult, RankingSummaryVariables>(RANKING_SUMMARY, {
        variables,
        fetchPolicy: 'cache-first',
        pollInterval: 5000,
    })

export const BUMP_DERIVATIVE_GRAPH_KEY = gql`
    mutation BumpDerivativeGraphKey {
        bumpDerivativeGraphKey {
            alwaysNil
        }
    }
`

export const DELETE_RANKING_PROGRESS = gql`
    mutation DeleteRankingProgress($graphKey: String!) {
        deleteRankingProgress(graphKey: $graphKey) {
            alwaysNil
        }
    }
`
