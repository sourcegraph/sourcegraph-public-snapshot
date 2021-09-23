import { ApolloError, FetchResult, MutationFunctionOptions, ApolloCache, OperationVariables } from '@apollo/client'

import { gql, useQuery, useMutation } from '@sourcegraph/shared/src/graphql/graphql'

import {
    Exact,
    CodeIntelligenceConfigurationPolicyFields,
    CodeIntelligenceConfigurationPoliciesResult,
    DeleteCodeIntelligenceConfigurationPolicyResult,
    DeleteCodeIntelligenceConfigurationPolicyVariables,
    IndexConfigurationResult,
    InferredIndexConfigurationResult,
    UpdateRepositoryIndexConfigurationResult,
    CodeIntelligenceConfigurationPolicyResult,
    GitObjectType,
    UpdateCodeIntelligenceConfigurationPolicyResult,
} from '../../../graphql-operations'

// Query
const defaultCodeIntelligenceConfigurationPolicyFieldsFragment = gql`
    fragment CodeIntelligenceConfigurationPolicyFields on CodeIntelligenceConfigurationPolicy {
        __typename
        id
        name
        type
        pattern
        protected
        retentionEnabled
        retentionDurationHours
        retainIntermediateCommits
        indexingEnabled
        indexCommitMaxAgeHours
        indexIntermediateCommits
    }
`

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

interface UseRepositoryConfigResult {
    configuration: string
    loadingRepository: boolean
    repositoryError: ApolloError | undefined
}

export const REPOSITORY_CONFIGURATION = gql`
    query IndexConfiguration($id: ID!) {
        node(id: $id) {
            ...RepositoryIndexConfigurationFields
        }
    }

    fragment RepositoryIndexConfigurationFields on Repository {
        __typename
        indexConfiguration {
            configuration
        }
    }
`

export const useRepositoryConfig = (id: string): UseRepositoryConfigResult => {
    const { data, loading, error } = useQuery<IndexConfigurationResult>(REPOSITORY_CONFIGURATION, {
        variables: { id },
    })

    const configuration = data?.node?.indexConfiguration?.configuration || ''

    return {
        configuration,
        loadingRepository: loading,
        repositoryError: error,
    }
}

export const INFERRED_CONFIGURATION = gql`
    query InferredIndexConfiguration($id: ID!) {
        node(id: $id) {
            ...RepositoryInferredIndexConfigurationFields
        }
    }

    fragment RepositoryInferredIndexConfigurationFields on Repository {
        __typename
        indexConfiguration {
            inferredConfiguration
        }
    }
`

interface UseInferredConfigResult {
    inferredConfiguration: string
    loadingInferred: boolean
    inferredError: ApolloError | undefined
}

export const useInferredConfig = (id: string): UseInferredConfigResult => {
    const { data, loading, error } = useQuery<InferredIndexConfigurationResult>(INFERRED_CONFIGURATION, {
        variables: { id },
    })

    const inferredConfiguration = data?.node?.indexConfiguration?.inferredConfiguration || ''

    return {
        inferredConfiguration,
        loadingInferred: loading,
        inferredError: error,
    }
}

export const POLICY_CONFIGURATION_BY_ID = gql`
    query CodeIntelligenceConfigurationPolicy($id: ID!) {
        node(id: $id) {
            ...CodeIntelligenceConfigurationPolicyFields
        }
    }

    ${defaultCodeIntelligenceConfigurationPolicyFieldsFragment}
`

interface UsePolicyConfigResult {
    policyConfig: CodeIntelligenceConfigurationPolicyFields | undefined
    loadingPolicyConfig: boolean
    policyConfigError: ApolloError | undefined
}

const emptyPolicy: CodeIntelligenceConfigurationPolicyFields = {
    __typename: 'CodeIntelligenceConfigurationPolicy',
    id: '',
    name: '',
    type: GitObjectType.GIT_COMMIT,
    pattern: '',
    protected: false,
    retentionEnabled: false,
    retentionDurationHours: null,
    retainIntermediateCommits: false,
    indexingEnabled: false,
    indexCommitMaxAgeHours: null,
    indexIntermediateCommits: false,
}

export const usePolicyConfigurationByID = (id: string): UsePolicyConfigResult => {
    const { data, loading, error } = useQuery<CodeIntelligenceConfigurationPolicyResult>(POLICY_CONFIGURATION_BY_ID, {
        variables: { id },
        skip: id === 'new',
    })

    const response = data?.node || undefined
    const isNew = id === 'new'

    return {
        policyConfig: isNew ? emptyPolicy : response,
        loadingPolicyConfig: loading,
        policyConfigError:
            !response && !isNew
                ? new ApolloError({ errorMessage: 'No such CodeIntelligenceConfigurationPolicy' })
                : error,
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

const UPDATE_CONFIGURATION_FOR_REPOSITORY = gql`
    mutation UpdateRepositoryIndexConfiguration($id: ID!, $content: String!) {
        updateRepositoryIndexConfiguration(repository: $id, configuration: $content) {
            alwaysNil
        }
    }
`

type UpdateConfigurationResult = Promise<
    FetchResult<UpdateRepositoryIndexConfigurationResult, Record<string, string>, Record<string, string>>
>
interface UpdateConfigurationForRepositoryResult {
    updateConfigForRepository: (
        options?: MutationFunctionOptions<UpdateRepositoryIndexConfigurationResult, OperationVariables> | undefined
    ) => UpdateConfigurationResult
    isUpdating: boolean
    updatingError: ApolloError | undefined
}

export const useUpdateConfigurationForRepository = (): UpdateConfigurationForRepositoryResult => {
    const [updateConfigForRepository, { loading, error }] = useMutation<UpdateRepositoryIndexConfigurationResult>(
        UPDATE_CONFIGURATION_FOR_REPOSITORY
    )

    return {
        updateConfigForRepository,
        isUpdating: loading,
        updatingError: error,
    }
}

const CREATE_POLICY_CONFIGURATION = gql`
    mutation CreateCodeIntelligenceConfigurationPolicy(
        $repositoryId: ID
        $name: String!
        $type: GitObjectType!
        $pattern: String!
        $retentionEnabled: Boolean!
        $retentionDurationHours: Int
        $retainIntermediateCommits: Boolean!
        $indexingEnabled: Boolean!
        $indexCommitMaxAgeHours: Int
        $indexIntermediateCommits: Boolean!
    ) {
        createCodeIntelligenceConfigurationPolicy(
            repository: $repositoryId
            name: $name
            type: $type
            pattern: $pattern
            retentionEnabled: $retentionEnabled
            retentionDurationHours: $retentionDurationHours
            retainIntermediateCommits: $retainIntermediateCommits
            indexingEnabled: $indexingEnabled
            indexCommitMaxAgeHours: $indexCommitMaxAgeHours
            indexIntermediateCommits: $indexIntermediateCommits
        ) {
            id
        }
    }
`

const UPDATE_POLICY_CONFIGURATION = gql`
    mutation UpdateCodeIntelligenceConfigurationPolicy(
        $id: ID!
        $name: String!
        $type: GitObjectType!
        $pattern: String!
        $retentionEnabled: Boolean!
        $retentionDurationHours: Int
        $retainIntermediateCommits: Boolean!
        $indexingEnabled: Boolean!
        $indexCommitMaxAgeHours: Int
        $indexIntermediateCommits: Boolean!
    ) {
        updateCodeIntelligenceConfigurationPolicy(
            id: $id
            name: $name
            type: $type
            pattern: $pattern
            retentionEnabled: $retentionEnabled
            retentionDurationHours: $retentionDurationHours
            retainIntermediateCommits: $retainIntermediateCommits
            indexingEnabled: $indexingEnabled
            indexCommitMaxAgeHours: $indexCommitMaxAgeHours
            indexIntermediateCommits: $indexIntermediateCommits
        ) {
            alwaysNil
        }
    }
`

type SavePolicyConfigResult = Promise<
    FetchResult<UpdateCodeIntelligenceConfigurationPolicyResult, Record<string, string>, Record<string, string>>
>
interface SavePolicyConfigurationResult {
    savePolicyConfiguration: (
        options?:
            | MutationFunctionOptions<UpdateCodeIntelligenceConfigurationPolicyResult, OperationVariables>
            | undefined
    ) => SavePolicyConfigResult
    isSaving: boolean
    savingError: ApolloError | undefined
}

export const useSavePolicyConfiguration = (isNew: boolean): SavePolicyConfigurationResult => {
    const mutation = isNew ? CREATE_POLICY_CONFIGURATION : UPDATE_POLICY_CONFIGURATION
    const [savePolicyConfiguration, { loading, error }] = useMutation<UpdateCodeIntelligenceConfigurationPolicyResult>(
        mutation
    )

    return {
        savePolicyConfiguration,
        isSaving: loading,
        savingError: error,
    }
}
