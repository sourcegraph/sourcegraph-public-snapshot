import { MutationTuple } from '@apollo/client'

import { dataOrThrowErrors, gql, useMutation } from '@sourcegraph/http-client'

import { useConnection, UseConnectionResult } from '../../../components/FilteredConnection/hooks/useConnection'
import {
    ExecutorSecretFields,
    CreateBatchChangesCredentialResult,
    CreateBatchChangesCredentialVariables,
    DeleteBatchChangesCredentialResult,
    DeleteBatchChangesCredentialVariables,
    GlobalBatchChangesCodeHostsResult,
    GlobalBatchChangesCodeHostsVariables,
    Scalars,
    UserExecutorSecretsResult,
    UserExecutorSecretsVariables,
} from '../../../graphql-operations'

export const CREATE_BATCH_CHANGES_CREDENTIAL = gql`
    mutation CreateBatchChangesCredential(
        $user: ID
        $credential: String!
        $username: String
        $externalServiceKind: ExternalServiceKind!
        $externalServiceURL: String!
    ) {
        createBatchChangesCredential(
            user: $user
            credential: $credential
            username: $username
            externalServiceKind: $externalServiceKind
            externalServiceURL: $externalServiceURL
        ) {
            ...BatchChangesCredentialFields
        }
    }
`

export const useCreateBatchChangesCredential = (): MutationTuple<
    CreateBatchChangesCredentialResult,
    CreateBatchChangesCredentialVariables
> => useMutation(CREATE_BATCH_CHANGES_CREDENTIAL)

export const DELETE_EXECUTOR_SECRET = gql`
    mutation DeleteExecutorSecret($scope: ExecutorSecretScope!, $id: ID!) {
        deleteExecutorSecret(scope: $scope, id: $id) {
            alwaysNil
        }
    }
`

export const useDeleteExecutorSecret = (): MutationTuple<DeleteExecutorSecretResult, DeleteExecutorSecretVariables> =>
    useMutation(DELETE_EXECUTOR_SECRET)

const CODE_HOST_FIELDS_FRAGMENT = gql`
    fragment ExecutorSecretConnectionFields on ExecutorSecretConnection {
        totalCount
        pageInfo {
            hasNextPage
            endCursor
        }
        nodes {
            ...ExecutorSecretFields
        }
    }

    fragment ExecutorSecretFields on ExecutorSecret {
        key
        scope
        createdAt
        updatedAt
        creator {
            id
            username
            displayName
            url
        }
        namespace {
            id
            namespaceName
            url
        }
    }
`

export const USER_CODE_HOSTS = gql`
    query UserExecutorSecrets($user: ID!, $scope: ExecutorSecretScope!, $first: Int, $after: String) {
        node(id: $user) {
            __typename
            ... on User {
                executorSecrets(scope: $scope, first: $first, after: $after) {
                    ...ExecutorSecretConnectionFields
                }
            }
        }
    }

    ${CODE_HOST_FIELDS_FRAGMENT}
`

export const useUserBatchChangesCodeHostConnection = (user: Scalars['ID']): UseConnectionResult<ExecutorSecretFields> =>
    useConnection<UserExecutorSecretsResult, UserExecutorSecretsVariables, ExecutorSecretFields>({
        query: USER_CODE_HOSTS,
        variables: {
            user,
            after: null,
            first: 15,
        },
        options: {
            fetchPolicy: 'no-cache',
        },
        getConnection: result => {
            const { node } = dataOrThrowErrors(result)

            if (!node) {
                throw new Error('User not found')
            }
            if (node.__typename !== 'User') {
                throw new Error(`Node is a ${node.__typename}, not a User`)
            }

            return node.batchChangesCodeHosts
        },
    })

export const GLOBAL_EXECUTOR_SECRETS = gql`
    query GlobalExecutorSecrets($scope: ExecutorSecretScope!, $first: Int, $after: String) {
        executorSecrets(scope: $scope, first: $first, after: $after) {
            ...ExecutorSecretConnectionFields
        }
    }

    ${CODE_HOST_FIELDS_FRAGMENT}
`

export const useGlobalExecutorSecretsConnection = (
    scope: ExecutorSecretScope
): UseConnectionResult<ExecutorSecretFields> =>
    useConnection<GlobalExecutorSecretsResult, GlobalExecutorSecretsVariables, ExecutorSecretFields>({
        query: GLOBAL_EXECUTOR_SECRETS,
        variables: {
            after: null,
            first: 15,
            scope,
        },
        options: {
            useURL: true,
            fetchPolicy: 'no-cache',
        },
        getConnection: result => {
            const { executorSecrets } = dataOrThrowErrors(result)

            return executorSecrets
        },
    })
