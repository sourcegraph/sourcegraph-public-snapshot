import { gql } from '@sourcegraph/http-client'

/**
 * GraphQL query for the list of pending access requests.
 */
export const PENDING_ACCESS_REQUESTS_LIST = gql`
    query PendingAccessRequestsList($first: Int, $last: Int, $after: String, $before: String) {
        accessRequests(status: PENDING, first: $first, last: $last, after: $after, before: $before) {
            totalCount
            pageInfo {
                hasNextPage
                hasPreviousPage
                endCursor
                startCursor
            }
            nodes {
                id
                email
                name
                createdAt
                additionalInfo
                status
            }
        }
    }
`

/**
 * GraphQL mutation for rejecting an access request.
 */
export const REJECT_ACCESS_REQUEST = gql`
    mutation RejectAccessRequest($id: ID!) {
        setAccessRequestStatus(id: $id, status: REJECTED) {
            alwaysNil
        }
    }
`

/**
 * GraphQL query for checking if a username exists.
 */
export const DOES_USERNAME_EXIST = gql`
    query DoesUsernameExist($username: String!) {
        user(username: $username) {
            id
        }
    }
`

/**
 * GraphQL mutation for approving an access request.
 */
export const APPROVE_ACCESS_REQUEST = gql`
    mutation ApproveAccessRequest($id: ID!) {
        setAccessRequestStatus(id: $id, status: APPROVED) {
            alwaysNil
        }
    }
`

/**
 * GraphQL mutation for creating a user.
 */
export const ACCESS_REQUEST_CREATE_USER = gql`
    mutation AccessRequestCreateUser($username: String!, $email: String) {
        createUser(username: $username, email: $email, verifiedEmail: false) {
            resetPasswordURL
        }
    }
`

/**
 * GraphQL query for the count of pending access requests.
 */
export const ACCESS_REQUESTS_COUNT = gql`
    query AccessRequestsCount {
        accessRequests(status: PENDING) {
            totalCount
        }
    }
`
