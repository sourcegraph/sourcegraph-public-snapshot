import type { MutationTuple } from '@apollo/client'

import { dataOrThrowErrors, gql, useMutation } from '@sourcegraph/http-client'

import {
    useShowMorePagination,
    type UseShowMorePaginationResult,
} from '../../../components/FilteredConnection/hooks/useShowMorePagination'
import type {
    CreateExecutorSecretResult,
    CreateExecutorSecretVariables,
    DeleteExecutorSecretResult,
    DeleteExecutorSecretVariables,
    ExecutorSecretAccessLogFields,
    ExecutorSecretAccessLogsResult,
    ExecutorSecretAccessLogsVariables,
    ExecutorSecretFields,
    ExecutorSecretScope,
    OrgExecutorSecretsResult,
    OrgExecutorSecretsVariables,
    Scalars,
    UpdateExecutorSecretResult,
    UpdateExecutorSecretVariables,
    UserExecutorSecretsResult,
    UserExecutorSecretsVariables,
} from '../../../graphql-operations'

const EXECUTOR_SECRET_FIELDS = gql`
    fragment ExecutorSecretFields on ExecutorSecret {
        id
        key
        scope
        overwritesGlobalSecret
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

export const CREATE_EXECUTOR_SECRET = gql`
    mutation CreateExecutorSecret($scope: ExecutorSecretScope!, $namespace: ID, $key: String!, $value: String!) {
        createExecutorSecret(scope: $scope, namespace: $namespace, key: $key, value: $value) {
            ...ExecutorSecretFields
        }
    }

    ${EXECUTOR_SECRET_FIELDS}
`

export const useCreateExecutorSecret = (): MutationTuple<CreateExecutorSecretResult, CreateExecutorSecretVariables> =>
    useMutation(CREATE_EXECUTOR_SECRET)

export const UPDATE_EXECUTOR_SECRET = gql`
    mutation UpdateExecutorSecret($scope: ExecutorSecretScope!, $secret: ID!, $value: String!) {
        updateExecutorSecret(scope: $scope, id: $secret, value: $value) {
            ...ExecutorSecretFields
        }
    }

    ${EXECUTOR_SECRET_FIELDS}
`

export const useUpdateExecutorSecret = (): MutationTuple<UpdateExecutorSecretResult, UpdateExecutorSecretVariables> =>
    useMutation(UPDATE_EXECUTOR_SECRET)

export const DELETE_EXECUTOR_SECRET = gql`
    mutation DeleteExecutorSecret($scope: ExecutorSecretScope!, $id: ID!) {
        deleteExecutorSecret(scope: $scope, id: $id) {
            alwaysNil
        }
    }
`

export const useDeleteExecutorSecret = (): MutationTuple<DeleteExecutorSecretResult, DeleteExecutorSecretVariables> =>
    useMutation(DELETE_EXECUTOR_SECRET)

const EXECUTOR_SECRET_CONNECTION_FIELDS = gql`
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

    ${EXECUTOR_SECRET_FIELDS}
`

export const USER_EXECUTOR_SECRETS = gql`
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

    ${EXECUTOR_SECRET_CONNECTION_FIELDS}
`

export const userExecutorSecretsConnectionFactory = (
    user: Scalars['ID'],
    scope: ExecutorSecretScope
): UseShowMorePaginationResult<UserExecutorSecretsResult, ExecutorSecretFields> =>
    // Scope has to be injected dynamically.
    // eslint-disable-next-line react-hooks/rules-of-hooks
    useShowMorePagination<UserExecutorSecretsResult, UserExecutorSecretsVariables, ExecutorSecretFields>({
        query: USER_EXECUTOR_SECRETS,
        variables: {
            user,
            scope,
        },
        options: {
            fetchPolicy: 'network-only',
        },
        getConnection: result => {
            const { node } = dataOrThrowErrors(result)

            if (!node) {
                throw new Error('User not found')
            }
            if (node.__typename !== 'User') {
                throw new Error(`Node is a ${node.__typename}, not a User`)
            }

            return node.executorSecrets
        },
    })

export const ORG_EXECUTOR_SECRETS = gql`
    query OrgExecutorSecrets($org: ID!, $scope: ExecutorSecretScope!, $first: Int, $after: String) {
        node(id: $org) {
            __typename
            ... on Org {
                executorSecrets(scope: $scope, first: $first, after: $after) {
                    ...ExecutorSecretConnectionFields
                }
            }
        }
    }

    ${EXECUTOR_SECRET_CONNECTION_FIELDS}
`

export const orgExecutorSecretsConnectionFactory = (
    org: Scalars['ID'],
    scope: ExecutorSecretScope
): UseShowMorePaginationResult<OrgExecutorSecretsResult, ExecutorSecretFields> =>
    // Scope has to be injected dynamically.
    // eslint-disable-next-line react-hooks/rules-of-hooks
    useShowMorePagination<OrgExecutorSecretsResult, OrgExecutorSecretsVariables, ExecutorSecretFields>({
        query: ORG_EXECUTOR_SECRETS,
        variables: {
            org,
            scope,
        },
        options: {
            fetchPolicy: 'network-only',
        },
        getConnection: result => {
            const { node } = dataOrThrowErrors(result)

            if (!node) {
                throw new Error('Org not found')
            }
            if (node.__typename !== 'Org') {
                throw new Error(`Node is a ${node.__typename}, not an Org`)
            }

            return node.executorSecrets
        },
    })

export const GLOBAL_EXECUTOR_SECRETS = gql`
    query GlobalExecutorSecrets($scope: ExecutorSecretScope!, $first: Int, $after: String) {
        executorSecrets(scope: $scope, first: $first, after: $after) {
            ...ExecutorSecretConnectionFields
        }
    }

    ${EXECUTOR_SECRET_CONNECTION_FIELDS}
`

export const EXECUTOR_SECRET_ACCESS_LOGS = gql`
    query ExecutorSecretAccessLogs($secret: ID!, $first: Int, $after: String) {
        node(id: $secret) {
            __typename
            ... on ExecutorSecret {
                accessLogs(first: $first, after: $after) {
                    ...ExecutorSecretAccessLogConnectionFields
                }
            }
        }
    }

    fragment ExecutorSecretAccessLogConnectionFields on ExecutorSecretAccessLogConnection {
        nodes {
            ...ExecutorSecretAccessLogFields
        }
        totalCount
        pageInfo {
            endCursor
            hasNextPage
        }
    }

    fragment ExecutorSecretAccessLogFields on ExecutorSecretAccessLog {
        id
        user {
            id
            email
            username
            displayName
            url
        }
        createdAt
    }
`

export const useExecutorSecretAccessLogsConnection = (
    secret: Scalars['ID']
): UseShowMorePaginationResult<ExecutorSecretAccessLogsResult, ExecutorSecretAccessLogFields> =>
    useShowMorePagination<
        ExecutorSecretAccessLogsResult,
        ExecutorSecretAccessLogsVariables,
        ExecutorSecretAccessLogFields
    >({
        query: EXECUTOR_SECRET_ACCESS_LOGS,
        variables: {
            secret,
        },
        options: {
            fetchPolicy: 'network-only',
        },
        getConnection: result => {
            const { node } = dataOrThrowErrors(result)

            if (!node) {
                throw new Error('Secret not found')
            }
            if (node.__typename !== 'ExecutorSecret') {
                throw new Error(`Node is a ${node.__typename}, not an ExecutorSecret`)
            }

            return node.accessLogs
        },
    })
