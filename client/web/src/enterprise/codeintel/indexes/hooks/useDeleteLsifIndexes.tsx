import { ApolloError, MutationFunctionOptions, FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import { DeleteLsifIndexesResult, DeleteLsifIndexesVariables } from '../../../../graphql-operations'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type DeleteLsifIndexesResults = Promise<FetchResult<DeleteLsifIndexesResult, Record<string, any>, Record<string, any>>>

interface UseDeleteLsifIndexesResult {
    handleDeleteLsifIndexes: (
        options?: MutationFunctionOptions<DeleteLsifIndexesResult, DeleteLsifIndexesVariables> | undefined
    ) => DeleteLsifIndexesResults
    deletesError: ApolloError | undefined
}

const DELETE_LSIF_INDEXES = gql`
    mutation DeleteLsifIndexes($query: String, $state: LSIFIndexState, $repository: ID) {
        deleteLSIFIndexes(query: $query, state: $state, repository: $repository) {
            alwaysNil
        }
    }
`

export const useDeleteLsifIndexes = (): UseDeleteLsifIndexesResult => {
    const [handleDeleteLsifIndexes, { error }] = useMutation<DeleteLsifIndexesResult, DeleteLsifIndexesVariables>(
        getDocumentNode(DELETE_LSIF_INDEXES)
    )

    return {
        handleDeleteLsifIndexes,
        deletesError: error,
    }
}
