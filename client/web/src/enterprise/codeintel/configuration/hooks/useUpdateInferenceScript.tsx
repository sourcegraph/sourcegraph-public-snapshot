import type { ApolloError, FetchResult, MutationFunctionOptions, OperationVariables } from '@apollo/client'

import { gql, useMutation } from '@sourcegraph/http-client'

import type { UpdateCodeGraphInferenceScriptResult } from '../../../../graphql-operations'

const UPDATE_INFERENCE_SCRIPT = gql`
    mutation UpdateCodeGraphInferenceScript($script: String!) {
        updateCodeIntelligenceInferenceScript(script: $script) {
            alwaysNil
        }
    }
`

type UpdateInferenceScriptFetchResult = Promise<
    FetchResult<UpdateCodeGraphInferenceScriptResult, Record<string, string>, Record<string, string>>
>
interface UpdateInferenceScriptResult {
    updateInferenceScript: (
        options?: MutationFunctionOptions<UpdateCodeGraphInferenceScriptResult, OperationVariables> | undefined
    ) => UpdateInferenceScriptFetchResult
    isUpdating: boolean
    updatingError: ApolloError | undefined
}

export const useUpdateInferenceScript = (): UpdateInferenceScriptResult => {
    const [updateInferenceScript, { loading, error }] =
        useMutation<UpdateCodeGraphInferenceScriptResult>(UPDATE_INFERENCE_SCRIPT)

    return {
        updateInferenceScript,
        isUpdating: loading,
        updatingError: error,
    }
}
