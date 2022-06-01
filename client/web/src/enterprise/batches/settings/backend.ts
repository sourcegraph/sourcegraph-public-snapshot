import { MutationTuple } from '@apollo/client'

import { dataOrThrowErrors, gql, useMutation } from '@sourcegraph/http-client'

import { useConnection, UseConnectionResult } from '../../../components/FilteredConnection/hooks/useConnection'
import {
    BatchChangesCodeHostFields,
    CreateBatchChangesCredentialResult,
    CreateBatchChangesCredentialVariables,
    DeleteBatchChangesCredentialResult,
    DeleteBatchChangesCredentialVariables,
    GlobalBatchChangesCodeHostsResult,
    GlobalBatchChangesCodeHostsVariables,
    Scalars,
    UserBatchChangesCodeHostsResult,
    UserBatchChangesCodeHostsVariables,
} from '../../../graphql-operations'

export const CREDENTIAL_FIELDS_FRAGMENT = gql`
    fragment BatchChangesCredentialFields on BatchChangesCredential {
        id
        sshPublicKey
        isSiteCredential
    }
`

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

    ${CREDENTIAL_FIELDS_FRAGMENT}
`

export const useCreateBatchChangesCredential = (): MutationTuple<
    CreateBatchChangesCredentialResult,
    CreateBatchChangesCredentialVariables
> => useMutation(CREATE_BATCH_CHANGES_CREDENTIAL)

export const DELETE_BATCH_CHANGES_CREDENTIAL = gql`
    mutation DeleteBatchChangesCredential($id: ID!) {
        deleteBatchChangesCredential(batchChangesCredential: $id) {
            alwaysNil
        }
    }
`

export const useDeleteBatchChangesCredential = (): MutationTuple<
    DeleteBatchChangesCredentialResult,
    DeleteBatchChangesCredentialVariables
> => useMutation(DELETE_BATCH_CHANGES_CREDENTIAL)

const CODE_HOST_FIELDS_FRAGMENT = gql`
    fragment BatchChangesCodeHostsFields on BatchChangesCodeHostConnection {
        totalCount
        pageInfo {
            hasNextPage
            endCursor
        }
        nodes {
            ...BatchChangesCodeHostFields
        }
    }

    fragment BatchChangesCodeHostFields on BatchChangesCodeHost {
        externalServiceKind
        externalServiceURL
        requiresSSH
        requiresUsername
        credential {
            ...BatchChangesCredentialFields
        }
    }

    ${CREDENTIAL_FIELDS_FRAGMENT}
`

export const USER_CODE_HOSTS = gql`
    query UserBatchChangesCodeHosts($user: ID!, $first: Int, $after: String) {
        node(id: $user) {
            __typename
            ... on User {
                batchChangesCodeHosts(first: $first, after: $after) {
                    ...BatchChangesCodeHostsFields
                }
            }
        }
    }

    ${CODE_HOST_FIELDS_FRAGMENT}
`

export const useUserBatchChangesCodeHostConnection = (
    user: Scalars['ID']
): UseConnectionResult<BatchChangesCodeHostFields> =>
    useConnection<UserBatchChangesCodeHostsResult, UserBatchChangesCodeHostsVariables, BatchChangesCodeHostFields>({
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

export const GLOBAL_CODE_HOSTS = gql`
    query GlobalBatchChangesCodeHosts($first: Int, $after: String) {
        batchChangesCodeHosts(first: $first, after: $after) {
            ...BatchChangesCodeHostsFields
        }
    }

    ${CODE_HOST_FIELDS_FRAGMENT}
`

export const useGlobalBatchChangesCodeHostConnection = (): UseConnectionResult<BatchChangesCodeHostFields> =>
    useConnection<GlobalBatchChangesCodeHostsResult, GlobalBatchChangesCodeHostsVariables, BatchChangesCodeHostFields>({
        query: GLOBAL_CODE_HOSTS,
        variables: {
            after: null,
            first: 15,
        },
        options: {
            useURL: true,
            fetchPolicy: 'no-cache',
        },
        getConnection: result => {
            const { batchChangesCodeHosts } = dataOrThrowErrors(result)

            return batchChangesCodeHosts
        },
    })
