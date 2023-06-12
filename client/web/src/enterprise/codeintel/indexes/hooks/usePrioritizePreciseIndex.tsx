import { ApolloError, MutationFunctionOptions, FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import { PrioritizePreciseIndexResult, PrioritizePreciseIndexVariables, Exact } from '../../../../graphql-operations'

type PrioritizePreciseIndexResults = Promise<
    FetchResult<PrioritizePreciseIndexResult, Record<string, any>, Record<string, any>>
>

interface UsePrioritizePreciseIndexResult {
    handlePrioritizePreciseIndex: (
        options?:
            | MutationFunctionOptions<
                  PrioritizePreciseIndexResult,
                  Exact<{
                      id: string
                  }>
              >
            | undefined
    ) => PrioritizePreciseIndexResults
    prioritizeError: ApolloError | undefined
}

const PRIORITIZE_PRECISE_INDEX = gql`
    mutation PrioritizePreciseIndex($id: ID!) {
        prioritizePreciseIndex(id: $id) {
            alwaysNil
        }
    }
`

export const usePrioritizePreciseIndex = (): UsePrioritizePreciseIndexResult => {
    const [handlePrioritizePreciseIndex, { error }] = useMutation<
        PrioritizePreciseIndexResult,
        PrioritizePreciseIndexVariables
    >(getDocumentNode(PRIORITIZE_PRECISE_INDEX))

    return {
        handlePrioritizePreciseIndex,
        prioritizeError: error,
    }
}
