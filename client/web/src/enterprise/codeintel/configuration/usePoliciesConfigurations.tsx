import { ApolloError, FetchResult, MutationFunctionOptions, ApolloCache } from '@apollo/client'

import { gql, useQuery, useMutation } from '@sourcegraph/shared/src/graphql/graphql'

import {
    Exact,
    CodeIntelligenceConfigurationPolicyFields,
    CodeIntelligenceConfigurationPoliciesResult,
    DeleteCodeIntelligenceConfigurationPolicyResult,
    DeleteCodeIntelligenceConfigurationPolicyVariables,
} from '../../../graphql-operations'

import { codeIntelligenceConfigurationPolicyFieldsFragment as defaultCodeIntelligenceConfigurationPolicyFieldsFragment } from './backend'

// Query
interface UsePoliciesConfigResult {
    policies: CodeIntelligenceConfigurationPolicyFields[]
    loadingPolicies: boolean
    policiesError: ApolloError | undefined
}

export const POLICIES_CONFIGURATION = gql`
    query CodeIntelligenceConfigurationPolicies($repositoryId: ID) {
        codeIntelligenceConfigurationPolicies(repository: $repositoryId) {
            ...CodeIntelligenceConfigurationPolicyFields
        }
    }

    ${defaultCodeIntelligenceConfigurationPolicyFieldsFragment}
`

export const usePoliciesConfig = (repositoryId?: string | null): UsePoliciesConfigResult => {
    const vars = repositoryId ? { variables: { repositoryId } } : {}
    const { data, error, loading } = useQuery<CodeIntelligenceConfigurationPoliciesResult>(POLICIES_CONFIGURATION, vars)

    return {
        policies: data?.codeIntelligenceConfigurationPolicies || [],
        loadingPolicies: loading,
        policiesError: error,
    }
}

// Mutations
export type DeletePolicyResult = Promise<
    | FetchResult<DeleteCodeIntelligenceConfigurationPolicyResult, Record<string, string>, Record<string, string>>
    | undefined
>

export interface UseDeletePoliciesResult {
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

export interface CachedRepositoryPolicies {
    __ref: string
}

export const updateDeletePolicyCache = (
    cache: ApolloCache<DeleteCodeIntelligenceConfigurationPolicyResult>,
    id: string
): boolean => {
    const policyReference = cache.identify({ __typename: 'CodeIntelligenceConfigurationPolicy', id })
    return cache.modify({
        fields: {
            codeIntelligenceConfigurationPolicies(existingPolicies: CachedRepositoryPolicies[] = []) {
                return existingPolicies.filter(({ __ref }) => __ref !== policyReference)
            },
        },
    })
}
