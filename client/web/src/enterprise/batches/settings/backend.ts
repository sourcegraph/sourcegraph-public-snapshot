import { gql } from '@sourcegraph/shared/src/graphql/graphql'

export const batchChangesCredentialFieldsFragment = gql`
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
`

export const DELETE_BATCH_CHANGES_CREDENTIAL = gql`
    mutation DeleteBatchChangesCredential($id: ID!) {
        deleteBatchChangesCredential(batchChangesCredential: $id) {
            alwaysNil
        }
    }
`

const batchChangesCodeHostsFieldsFragment = gql`
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

export const USER_BATCH_CHANGES_CODE_HOSTS = gql`
    query UserBatchChangesCodeHosts($user: ID!, $first: Int, $after: String) {
        node(id: $user) {
            __typename
            ... on User {
                id
                batchChangesCodeHosts(first: $first, after: $after) {
                    ...BatchChangesCodeHostsFields
                }
            }
        }
    }

    ${batchChangesCodeHostsFieldsFragment}
`

export const GLOBAL_BATCH_CHANGES_CODE_HOSTS = gql`
    query GlobalBatchChangesCodeHosts($first: Int, $after: String) {
        batchChangesCodeHosts(first: $first, after: $after) {
            ...BatchChangesCodeHostsFields
        }
    }

    ${batchChangesCodeHostsFieldsFragment}
`
