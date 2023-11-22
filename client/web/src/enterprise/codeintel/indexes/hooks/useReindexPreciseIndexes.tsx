import { type ApolloError, type FetchResult, type MutationFunctionOptions, useMutation } from '@apollo/client'

import { getDocumentNode, gql } from '@sourcegraph/http-client'

import type { ReindexPreciseIndexesResult, ReindexPreciseIndexesVariables } from '../../../../graphql-operations'

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
        $indexerKey: String
        $repo: ID
        $isLatestForRepo: Boolean
    ) {
        reindexPreciseIndexes(
            query: $query
            states: $states
            indexerKey: $indexerKey
            repository: $repo
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
                repo: options?.variables?.repo ?? null,
                query: options?.variables?.query ?? null,
                states: options?.variables?.states ?? null,
                indexerKey: options?.variables?.indexerKey ?? null,
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
