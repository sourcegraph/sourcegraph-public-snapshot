import { ApolloError, FetchResult, MutationFunctionOptions, useMutation } from '@apollo/client'

import { getDocumentNode, gql } from '@sourcegraph/http-client'

import { ReindexPreciseIndexesResult, ReindexPreciseIndexesVariables } from '../../../../graphql-operations'

type ReindexPreciseIndexesResults = Promise<
    FetchResult<ReindexPreciseIndexesResult, Record<string, any>, Record<string, any>>
>

interface UseReindexPreciseIndexesResult {
    handleReindexPreciseIndexes: (
        options?: MutationFunctionOptions<ReindexPreciseIndexesResult, ReindexPreciseIndexesVariables> | undefined
    ) => ReindexPreciseIndexesResults
    reindexesError: ApolloError | undefined
}

const REINDEX_PRECISE_INDEXES = gql`
    mutation ReindexPreciseIndexes(
        $query: String
        $states: [PreciseIndexState!]
        $repository: ID
        $isLatestForRepo: Boolean
    ) {
        reindexPreciseIndexes(
            query: $query
            states: $states
            repository: $repository
            isLatestForRepo: $isLatestForRepo
        ) {
            alwaysNil
        }
    }
`

export const useReindexPreciseIndexes = (): UseReindexPreciseIndexesResult => {
    const [handleReindexPreciseIndexes, { error }] = useMutation<
        ReindexPreciseIndexesResult,
        ReindexPreciseIndexesVariables
    >(getDocumentNode(REINDEX_PRECISE_INDEXES))

    return {
        handleReindexPreciseIndexes: (
            options?: MutationFunctionOptions<ReindexPreciseIndexesResult, ReindexPreciseIndexesVariables> | undefined
        ): ReindexPreciseIndexesResults => {
            const variables = {
                query: options?.variables?.query ?? null,
                states: options?.variables?.states ?? null,
                repository: options?.variables?.repository ?? null,
                isLatestForRepo: options?.variables?.isLatestForRepo ?? null,
            }

            return handleReindexPreciseIndexes({
                ...options,
                variables,
            })
        },
        reindexesError: error,
    }
}
