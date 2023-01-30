import { ApolloError, MutationFunctionOptions, FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import { ReindexPreciseIndexesResult, ReindexPreciseIndexesVariables } from '../../../../graphql-operations'
 
type ReindexPreciseIndexesResults = Promise<
    FetchResult<ReindexPreciseIndexesResult, Record<string, any>, Record<string, any>>
>

interface UseReindexPreciseIndexesResult {
    handleReindexPreciseIndexes: (
        options?: MutationFunctionOptions<ReindexPreciseIndexesResult, ReindexPreciseIndexesVariables> | undefined
    ) => ReindexPrecisesIndexesResults
    reindexesError: ApolloError | undefined
}

const REINDEX_PRECISE_INDEXES = gql`
    mutation ReindexPreciseIndexes($query: String, $state: PreciseIndexState, $repository: ID) {
        reindexPreciseIndexes(query: $query, state: $state, repository: $repository) {
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
        handleReindexPreciseIndexes,
        reindexesError: error,
    }
}
