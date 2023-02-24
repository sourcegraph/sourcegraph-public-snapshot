import { ApolloError, FetchResult, MutationFunctionOptions, useMutation } from '@apollo/client'

import { getDocumentNode, gql } from '@sourcegraph/http-client'

import { Exact, ReindexPreciseIndexResult, ReindexPreciseIndexVariables } from '../../../../graphql-operations'

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
