import { type ApolloError, type FetchResult, type MutationFunctionOptions, useMutation } from '@apollo/client'

import { getDocumentNode, gql } from '@sourcegraph/http-client'

import type { DeletePreciseIndexesResult, DeletePreciseIndexesVariables } from '../../../../graphql-operations'

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
        $states: [PreciseIndexState!]
        $indexerKey: String
        $repo: ID
        $isLatestForRepo: Boolean
    ) {
        deletePreciseIndexes(
            query: $query
            states: $states
            indexerKey: $indexerKey
            repository: $repo
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
            options?: MutationFunctionOptions<DeletePreciseIndexesResult, DeletePreciseIndexesVariables> | undefined
        ): DeletePreciseIndexesResults => {
            const variables = {
                repo: options?.variables?.repo ?? null,
                query: options?.variables?.query ?? null,
                states: options?.variables?.states ?? null,
                indexerKey: options?.variables?.indexerKey ?? null,
                isLatestForRepo: options?.variables?.isLatestForRepo ?? null,
            }

            return handleDeletePreciseIndexes({ ...options, variables })
        },
        deletesError: error,
    }
}
