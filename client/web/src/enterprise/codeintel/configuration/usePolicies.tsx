import { ApolloClient, ApolloError, FetchResult, MutationFunctionOptions, OperationVariables } from '@apollo/client'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { getDocumentNode, gql, useMutation, useQuery } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import {
    CodeIntelligenceConfigurationPoliciesResult,
    CodeIntelligenceConfigurationPoliciesVariables,
    CodeIntelligenceConfigurationPolicyFields,
    CodeIntelligenceConfigurationPolicyResult,
    DeleteCodeIntelligenceConfigurationPolicyResult,
    DeleteCodeIntelligenceConfigurationPolicyVariables,
    Exact,
    GitObjectType,
    IndexConfigurationResult,
    InferredIndexConfigurationResult,
    UpdateCodeIntelligenceConfigurationPolicyResult,
    UpdateRepositoryIndexConfigurationResult,
} from '../../../graphql-operations'

// Query
const defaultCodeIntelligenceConfigurationPolicyFieldsFragment = gql`
    fragment CodeIntelligenceConfigurationPolicyFields on CodeIntelligenceConfigurationPolicy {
        __typename
        id
        name
        repository {
            id
            name
        }
        repositoryPatterns
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

export const nullPolicy = {
    __typename: 'CodeIntelligenceConfigurationPolicy' as const,
    id: '',
    name: '',
    repositoryPatterns: null,
    type: GitObjectType.GIT_UNKNOWN,
    pattern: '',
    protected: false,
    retentionEnabled: false,
    retentionDurationHours: null,
    retainIntermediateCommits: false,
    indexingEnabled: false,
    indexCommitMaxAgeHours: null,
    indexIntermediateCommits: false,
    repository: null,
}

interface PolicyConnection {
    nodes: CodeIntelligenceConfigurationPolicyFields[]
    totalCount: number | null
    pageInfo: { endCursor: string | null; hasNextPage: boolean }
}

export const POLICIES_CONFIGURATION = gql`
    query CodeIntelligenceConfigurationPolicies(
        $repository: ID
        $query: String
        $forDataRetention: Boolean
        $forIndexing: Boolean
        $first: Int
        $after: String
    ) {
        codeIntelligenceConfigurationPolicies(
            repository: $repository
            query: $query
            forDataRetention: $forDataRetention
            forIndexing: $forIndexing
            first: $first
            after: $after
        ) {
            nodes {
                ...CodeIntelligenceConfigurationPolicyFields
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }

    ${defaultCodeIntelligenceConfigurationPolicyFieldsFragment}
`

export const queryPolicies = (
    {
        repository,
        first,
        query,
        forDataRetention,
        forIndexing,
        after,
    }: GQL.ICodeIntelligenceConfigurationPoliciesOnQueryArguments,
    client: ApolloClient<object>
): Observable<PolicyConnection> => {
    const vars: CodeIntelligenceConfigurationPoliciesVariables = {
        repository: repository ?? null,
        query: query ?? null,
        forDataRetention: forDataRetention ?? null,
        forIndexing: forIndexing ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<CodeIntelligenceConfigurationPoliciesResult>({
            query: getDocumentNode(POLICIES_CONFIGURATION),
            variables: vars,
        })
    ).pipe(
        map(({ data }) => data),
        map(({ codeIntelligenceConfigurationPolicies }) => codeIntelligenceConfigurationPolicies)
    )
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

    const configuration = (data?.node?.__typename === 'Repository' && data.node.indexConfiguration?.configuration) || ''

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

    const inferredConfiguration =
        (data?.node?.__typename === 'Repository' && data.node.indexConfiguration?.inferredConfiguration) || ''

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
    repository: null,
    repositoryPatterns: null,
    type: GitObjectType.GIT_UNKNOWN,
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

    const response = (data?.node?.__typename === 'CodeIntelligenceConfigurationPolicy' && data.node) || undefined
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
type DeletePolicyResult = Promise<
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
        $repositoryPatterns: [String!]
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
            repositoryPatterns: $repositoryPatterns
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
        $repositoryPatterns: [String!]
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
            repositoryPatterns: $repositoryPatterns
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
