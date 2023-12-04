import { type ApolloError, gql, useQuery } from '@apollo/client'

import type { CodeGraphInferenceScriptResult } from '../../../../graphql-operations'

interface UseInferenceScriptResult {
    inferenceScript: string
    loadingScript: boolean
    fetchError: ApolloError | undefined
}

export const INFERENCE_SCRIPT = gql`
    query CodeGraphInferenceScript {
        codeIntelligenceInferenceScript
    }
`

export const useInferenceScript = (): UseInferenceScriptResult => {
    const { data, loading, error } = useQuery<CodeGraphInferenceScriptResult>(INFERENCE_SCRIPT, {
        nextFetchPolicy: 'cache-first',
    })

    return {
        inferenceScript: data?.codeIntelligenceInferenceScript ?? '',
        loadingScript: loading,
        fetchError: error,
    }
}
