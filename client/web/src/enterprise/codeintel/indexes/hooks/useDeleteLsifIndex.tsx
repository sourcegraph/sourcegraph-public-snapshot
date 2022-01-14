import { ApolloError, MutationFunctionOptions, FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import { DeleteLsifIndexResult, DeleteLsifIndexVariables, Exact } from '../../../../graphql-operations'

type DeleteLsifIndexResults = Promise<FetchResult<DeleteLsifIndexResult, Record<string, any>, Record<string, any>>>

interface UseDeleteLsifIndexResult {
    handleDeleteLsifIndex: (
        options?:
            | MutationFunctionOptions<
                  DeleteLsifIndexResult,
                  Exact<{
                      id: string
                  }>
              >
            | undefined
    ) => DeleteLsifIndexResults
    deleteError: ApolloError | undefined
}

const DELETE_LSIF_INDEX = gql`
    mutation DeleteLsifIndex($id: ID!) {
        deleteLSIFIndex(id: $id) {
            alwaysNil
        }
    }
`

export const useDeleteLsifIndex = (): UseDeleteLsifIndexResult => {
    const [handleDeleteLsifIndex, { error }] = useMutation<DeleteLsifIndexResult, DeleteLsifIndexVariables>(
        getDocumentNode(DELETE_LSIF_INDEX)
    )

    return {
        handleDeleteLsifIndex,
        deleteError: error,
    }
}
