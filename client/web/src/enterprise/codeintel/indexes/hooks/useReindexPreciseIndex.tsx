import { type ApolloError, type MutationFunctionOptions, type FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import type { ReindexPreciseIndexResult, ReindexPreciseIndexVariables, Exact } from '../../../../graphql-operations'

type ReindexPreciseIndexResults = Promise<
    FetchResult<ReindexPreciseIndexResult, Record<string, any>, Record<string, any>>
>

interface UseReindexPreciseIndexResult {
    handleReindexPreciseIndex: (
        options?:
            | MutationFunctionOptions<
                  ReindexPreciseIndexResult,
                  Exact<{
                      id: string
                  }>
              >
            | undefined
    ) => ReindexPreciseIndexResults
    reindexError: ApolloError | undefined
}

const REINDEX_PRECISE_INDEX = gql`
    mutation ReindexPreciseIndex($id: ID!) {
        reindexPreciseIndex(id: $id) {
            alwaysNil
        }
    }
`

export const useReindexPreciseIndex = (): UseReindexPreciseIndexResult => {
    const [handleReindexPreciseIndex, { error }] = useMutation<ReindexPreciseIndexResult, ReindexPreciseIndexVariables>(
        getDocumentNode(REINDEX_PRECISE_INDEX)
    )

    return {
        handleReindexPreciseIndex,
        reindexError: error,
    }
}
