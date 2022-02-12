import { Observable } from 'rxjs'
import { map, mapTo } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../../backend/graphql'
import {
    BatchChangesCodeHostFields,
    BatchChangesCodeHostsFields,
    BatchChangesCredentialFields,
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
import { useConnection, UseConnectionResult } from '../../../components/FilteredConnection/hooks/useConnection'

export const batchChangesCredentialFieldsFragment = gql`
    fragment BatchChangesCredentialFields on BatchChangesCredential {
        id
        sshPublicKey
        isSiteCredential
    }
`

export function createBatchChangesCredential(
    args: CreateBatchChangesCredentialVariables
): Promise<BatchChangesCredentialFields> {
    return requestGraphQL<CreateBatchChangesCredentialResult, CreateBatchChangesCredentialVariables>(
        gql`
            mutation CreateBatchChangesCredential(
                $user: ID
                $credential: String!
                $externalServiceKind: ExternalServiceKind!
                $externalServiceURL: String!
            ) {
                createBatchChangesCredential(
                    user: $user
                    credential: $credential
                    externalServiceKind: $externalServiceKind
                    externalServiceURL: $externalServiceURL
                ) {
                    ...BatchChangesCredentialFields
                }
            }

            ${batchChangesCredentialFieldsFragment}
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.createBatchChangesCredential)
        )
        .toPromise()
}

export function deleteBatchChangesCredential(id: Scalars['ID']): Promise<void> {
    return requestGraphQL<DeleteBatchChangesCredentialResult, DeleteBatchChangesCredentialVariables>(
        gql`
            mutation DeleteBatchChangesCredential($id: ID!) {
                deleteBatchChangesCredential(batchChangesCredential: $id) {
                    alwaysNil
                }
            }
        `,
        { id }
    )
        .pipe(map(dataOrThrowErrors), mapTo(undefined))
        .toPromise()
}

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
        credential {
            ...BatchChangesCredentialFields
        }
    }

    ${batchChangesCredentialFieldsFragment}
`

const USER_CODE_HOSTS_QUERY = gql`
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
        query: USER_CODE_HOSTS_QUERY,
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
            if (!node.batchChangesCodeHosts) {
                throw new Error('No code hosts found')
            }

            return node.batchChangesCodeHosts
        },
    })

export const queryUserBatchChangesCodeHosts = ({
    user,
    first,
    after,
}: UserBatchChangesCodeHostsVariables): Observable<BatchChangesCodeHostsFields> =>
    requestGraphQL<UserBatchChangesCodeHostsResult, UserBatchChangesCodeHostsVariables>(gql``, {
        user,
        first,
        after,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (data.node === null) {
                throw new Error('User not found')
            }
            if (data.node.__typename !== 'User') {
                throw new Error(`Node is a ${data.node.__typename}, not a User`)
            }
            return data.node.batchChangesCodeHosts
        })
    )

const GLOBAL_CODE_HOSTS_QUERY = gql`
    query GlobalBatchChangesCodeHosts($first: Int, $after: String) {
        batchChangesCodeHosts(first: $first, after: $after) {
            ...BatchChangesCodeHostsFields
        }
    }

    ${CODE_HOST_FIELDS_FRAGMENT}
`

export const useGlobalBatchChangesCodeHostConnection = (): UseConnectionResult<BatchChangesCodeHostFields> =>
    useConnection<GlobalBatchChangesCodeHostsResult, GlobalBatchChangesCodeHostsVariables, BatchChangesCodeHostFields>({
        query: GLOBAL_CODE_HOSTS_QUERY,
        variables: {
            after: null,
            first: 20,
        },
        options: {
            useURL: true,
            fetchPolicy: 'no-cache',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            if (!data.batchChangesCodeHosts) {
                throw new Error('No code hosts found')
            }

            return data.batchChangesCodeHosts
        },
    })

export const queryGlobalBatchChangesCodeHosts = ({
    first,
    after,
}: GlobalBatchChangesCodeHostsVariables): Observable<BatchChangesCodeHostsFields> =>
    requestGraphQL<GlobalBatchChangesCodeHostsResult, GlobalBatchChangesCodeHostsVariables>(gql``, {
        first,
        after,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.batchChangesCodeHosts)
    )
