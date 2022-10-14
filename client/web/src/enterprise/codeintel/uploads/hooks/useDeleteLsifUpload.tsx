import { ApolloError, MutationFunctionOptions, FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import { DeleteLsifUploadResult, DeleteLsifUploadVariables, Exact } from '../../../../graphql-operations'

type DeleteLsifUploadResults = Promise<FetchResult<DeleteLsifUploadResult, Record<string, any>, Record<string, any>>>

interface UseDeleteLsifUploadResult {
    handleDeleteLsifUpload: (
        options?:
            | MutationFunctionOptions<
                  DeleteLsifUploadResult,
                  Exact<{
                      id: string
                  }>
              >
            | undefined
    ) => DeleteLsifUploadResults
    deleteError: ApolloError | undefined
}

const DELETE_LSIF_UPLOADS = gql`
    mutation DeleteLsifUpload($id: ID!) {
        deleteLSIFUpload(id: $id) {
            alwaysNil
        }
    }
`

export const useDeleteLsifUpload = (): UseDeleteLsifUploadResult => {
    const [handleDeleteLsifUpload, { error }] = useMutation<DeleteLsifUploadResult, DeleteLsifUploadVariables>(
        getDocumentNode(DELETE_LSIF_UPLOADS)
    )

    return {
        handleDeleteLsifUpload,
        deleteError: error,
    }
}
