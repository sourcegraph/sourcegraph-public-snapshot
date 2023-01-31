import { ApolloError, MutationFunctionOptions, FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import {
    DeletePreciseIndexesResult,
    DeletePreciseIndexesVariables,
    PreciseIndexState,
} from '../../../../graphql-operations'

type DeletePreciseIndexesResults = Promise<
    FetchResult<DeletePreciseIndexesResult, Record<string, any>, Record<string, any>>
>

interface UseDeletePreciseIndexesResult {
    handleDeletePreciseIndexes: (
        options?:
            | MutationFunctionOptions<
                  DeletePreciseIndexesResult,
                  Omit<DeletePreciseIndexesVariables, 'states'> & { state?: PreciseIndexState }
              >
            | undefined
    ) => DeletePreciseIndexesResults
    deletesError: ApolloError | undefined
}

const DELETE_PRECISE_INDEXES = gql`
    mutation DeletePreciseIndexes(
        $query: String
        $states: [PreciseIndexState!]
        $repository: ID
        $isLatestForRepo: Boolean
    ) {
        deletePreciseIndexes(
            query: $query
            states: $states
            repository: $repository
            isLatestForRepo: $isLatestForRepo
        ) {
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
        handleDeletePreciseIndexes: (
            options?:
                | MutationFunctionOptions<
                      DeletePreciseIndexesResult,
                      Omit<DeletePreciseIndexesVariables, 'states'> & { state?: PreciseIndexState }
                  >
                | undefined
        ): DeletePreciseIndexesResults => {
            const variables = {
                query: options?.variables?.query ?? null,
                states: options?.variables?.state ? [options.variables.state] : null,
                repository: options?.variables?.repository ?? null,
                isLatestForRepo: options?.variables?.isLatestForRepo ?? null,
            }

            return handleDeletePreciseIndexes({ ...options, variables })
        },
        deletesError: error,
    }
}
