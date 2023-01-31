import { ApolloError, FetchResult, MutationFunctionOptions, useMutation } from '@apollo/client'

import { getDocumentNode, gql } from '@sourcegraph/http-client'

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
    mutation DeletePreciseIndexes($query: String, $states: [PreciseIndexState!], $repo: ID, $isLatestForRepo: Boolean) {
        deletePreciseIndexes(query: $query, states: $states, repository: $repo, isLatestForRepo: $isLatestForRepo) {
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
            options?: MutationFunctionOptions<DeletePreciseIndexesResult, DeletePreciseIndexesVariables> | undefined
        ): DeletePreciseIndexesResults => {
            const variables = {
                repo: options?.variables?.repo ?? null,
                query: options?.variables?.query ?? null,
                states: options?.variables?.states ?? null,
                isLatestForRepo: options?.variables?.isLatestForRepo ?? null,
            }

            return handleDeletePreciseIndexes({ ...options, variables })
        },
        deletesError: error,
    }
}
