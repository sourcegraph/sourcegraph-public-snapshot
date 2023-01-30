import { ApolloError, MutationFunctionOptions, FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import { DeletePreciseIndexesResult, DeletePreciseIndexesVariables } from '../../../../graphql-operations'
 
type DeletePreciseIndexesResults = Promise<
    FetchResult<DeletePreciseIndexesResult, Record<string, any>, Record<string, any>>
>

interface UseDeletePreciseIndexesResult {
    handleDeletePreciseIndexes: (
        options?: MutationFunctionOptions<DeletePreciseIndexesResult, DeletePreciseIndexesVariables> | undefined
    ) => DeletePreciseIndexesResults
    deletesError: ApolloError | undefined
}

const DELETE_PRECISE_INDEXES = gql`
    mutation DeletePreciseIndexes(
        $query: String
        $state: PreciseIndexState
        $repository: ID
        $isLatestForRepo: Boolean
    ) {
        deletePreciseIndexes(query: $query, state: $state, repository: $repository, isLatestForRepo: $isLatestForRepo) {
            alwaysNil
        }
    }
`

export const useDeletePreciseIndexes = (): UseDeletePreciseIndexesResult => {
    const [handleDeletePreciseIndexes, { error }] = useMutation<
        DeletePreciseIndexesResult,
        DeletePreciseIndexesVariables
    >(getDocumentNode(DELETE_PRECISE_INDEXES))

    return {
        handleDeletePreciseIndexes,
        deletesError: error,
    }
}
