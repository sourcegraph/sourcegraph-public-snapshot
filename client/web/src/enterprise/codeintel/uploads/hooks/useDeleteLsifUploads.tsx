import { ApolloError, MutationFunctionOptions, FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import { DeleteLsifUploadsResult, DeleteLsifUploadsVariables } from '../../../../graphql-operations'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type DeleteLsifUploadsResults = Promise<FetchResult<DeleteLsifUploadsResult, Record<string, any>, Record<string, any>>>

interface UseDeleteLsifUploadsResult {
    handleDeleteLsifUploads: (
        options?: MutationFunctionOptions<DeleteLsifUploadsResult, DeleteLsifUploadsVariables> | undefined
    ) => DeleteLsifUploadsResults
    deletesError: ApolloError | undefined
}

const DELETE_LSIF_UPLOAD = gql`
    mutation DeleteLsifUploads($query: String, $state: LSIFUploadState, $isLatestForRepo: Boolean, $repository: ID) {
        deleteLSIFUploads(query: $query, state: $state, isLatestForRepo: $isLatestForRepo, repository: $repository) {
            alwaysNil
        }
    }
`

export const useDeleteLsifUploads = (): UseDeleteLsifUploadsResult => {
    const [handleDeleteLsifUploads, { error }] = useMutation<DeleteLsifUploadsResult, DeleteLsifUploadsVariables>(
        getDocumentNode(DELETE_LSIF_UPLOAD)
    )

    return {
        handleDeleteLsifUploads,
        deletesError: error,
    }
}
