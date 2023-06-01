import { ApolloError } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'

import { RankingSummaryResult, RankingSummaryVariables } from '../../../../graphql-operations'

export const RankingSummaryFieldsFragment = gql`
    fragment RankingSummaryFields on RankingSummary {
        graphKey
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
            ...RankingSummaryFields
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
} =>
    useQuery<RankingSummaryResult, RankingSummaryVariables>(RANKING_SUMMARY, {
        variables,
        fetchPolicy: 'cache-first',
    })

export const BUMP_DERIVATIVE_GRAPH_KEY = gql`
    mutation BumpDerivativeGraphKey {
        bumpDerivativeGraphKey {
            alwaysNil
        }
    }
`
