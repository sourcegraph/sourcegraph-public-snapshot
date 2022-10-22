import { ApolloError, MutationFunctionOptions, FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import { ReindexLsifIndexResult, ReindexLsifIndexVariables, Exact } from '../../../../graphql-operations'

type ReindexLsifIndexResults = Promise<FetchResult<ReindexLsifIndexResult, Record<string, any>, Record<string, any>>>

interface UseReindexLsifIndexResult {
    handleReindexLsifIndex: (
        options?:
            | MutationFunctionOptions<
                  ReindexLsifIndexResult,
                  Exact<{
                      id: string
                  }>
              >
            | undefined
    ) => ReindexLsifIndexResults
    reindexError: ApolloError | undefined
}

const REINDEX_LSIF_INDEX = gql`
    mutation ReindexLsifIndex($id: ID!) {
        reindexLSIFIndex(id: $id) {
            alwaysNil
        }
    }
`

export const useReindexLsifIndex = (): UseReindexLsifIndexResult => {
    const [handleReindexLsifIndex, { error }] = useMutation<ReindexLsifIndexResult, ReindexLsifIndexVariables>(
        getDocumentNode(REINDEX_LSIF_INDEX)
    )

    return {
        handleReindexLsifIndex,
        reindexError: error,
    }
}
