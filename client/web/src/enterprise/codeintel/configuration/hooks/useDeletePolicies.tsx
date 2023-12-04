import type { ApolloError, FetchResult, MutationFunctionOptions } from '@apollo/client'

import { gql, useMutation } from '@sourcegraph/http-client'

import type {
    DeleteCodeIntelligenceConfigurationPolicyResult,
    DeleteCodeIntelligenceConfigurationPolicyVariables,
    Exact,
} from '../../../../graphql-operations'

type DeletePolicyResult = Promise<
    | FetchResult<DeleteCodeIntelligenceConfigurationPolicyResult, Record<string, string>, Record<string, string>>
    | undefined
>

interface UseDeletePoliciesResult {
    handleDeleteConfig: (
        options?:
            | MutationFunctionOptions<
                  DeleteCodeIntelligenceConfigurationPolicyResult,
                  Exact<{
                      id: string
                  }>
              >
            | undefined
    ) => DeletePolicyResult
    isDeleting: boolean
    deleteError: ApolloError | undefined
}

const DELETE_POLICY_BY_ID = gql`
    mutation DeleteCodeIntelligenceConfigurationPolicy($id: ID!) {
        deleteCodeIntelligenceConfigurationPolicy(policy: $id) {
            alwaysNil
        }
    }
`

export const useDeletePolicies = (): UseDeletePoliciesResult => {
    const [handleDeleteConfig, { loading, error }] = useMutation<
        DeleteCodeIntelligenceConfigurationPolicyResult,
        DeleteCodeIntelligenceConfigurationPolicyVariables
    >(DELETE_POLICY_BY_ID)

    return {
        handleDeleteConfig,
        isDeleting: loading,
        deleteError: error,
    }
}
