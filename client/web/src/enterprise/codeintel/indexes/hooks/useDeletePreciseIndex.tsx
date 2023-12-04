import { type ApolloError, type MutationFunctionOptions, type FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import type { DeletePreciseIndexResult, DeletePreciseIndexVariables, Exact } from '../../../../graphql-operations'

type DeletePreciseIndexResults = Promise<
    FetchResult<DeletePreciseIndexResult, Record<string, any>, Record<string, any>>
>

interface UseDeletePreciseIndexResult {
    handleDeletePreciseIndex: (
        options?:
            | MutationFunctionOptions<
                  DeletePreciseIndexResult,
                  Exact<{
                      id: string
                  }>
              >
            | undefined
    ) => DeletePreciseIndexResults
    deleteError: ApolloError | undefined
}

const DELETE_PRECISE_INDEX = gql`
    mutation DeletePreciseIndex($id: ID!) {
        deletePreciseIndex(id: $id) {
            alwaysNil
        }
    }
`

export const useDeletePreciseIndex = (): UseDeletePreciseIndexResult => {
    const [handleDeletePreciseIndex, { error }] = useMutation<DeletePreciseIndexResult, DeletePreciseIndexVariables>(
        getDocumentNode(DELETE_PRECISE_INDEX)
    )

    return {
        handleDeletePreciseIndex,
        deleteError: error,
    }
}
