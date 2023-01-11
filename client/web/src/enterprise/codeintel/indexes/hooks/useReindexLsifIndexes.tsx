import { ApolloError, MutationFunctionOptions, FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import { ReindexLsifIndexesResult, ReindexLsifIndexesVariables } from '../../../../graphql-operations'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type ReindexLsifIndexesResults = Promise<
    FetchResult<ReindexLsifIndexesResult, Record<string, any>, Record<string, any>>
>

interface UseReindexLsifIndexesResult {
    handleReindexLsifIndexes: (
        options?: MutationFunctionOptions<ReindexLsifIndexesResult, ReindexLsifIndexesVariables> | undefined
    ) => ReindexLsifIndexesResults
    reindexesError: ApolloError | undefined
}

const REINDEX_LSIF_INDEXES = gql`
    mutation ReindexLsifIndexes($query: String, $state: LSIFIndexState, $repository: ID) {
        reindexLSIFIndexes(query: $query, state: $state, repository: $repository) {
            alwaysNil
        }
    }
`

export const useReindexLsifIndexes = (): UseReindexLsifIndexesResult => {
    const [handleReindexLsifIndexes, { error }] = useMutation<ReindexLsifIndexesResult, ReindexLsifIndexesVariables>(
        getDocumentNode(REINDEX_LSIF_INDEXES)
    )

    return {
        handleReindexLsifIndexes,
        reindexesError: error,
    }
}
