import { gql } from '@sourcegraph/http-client'

/**
 * GraphQL query for the list of pending access requests.
 */
export const GET_ACCESS_REQUESTS_LIST = gql`
    fragment AccessRequestNode on AccessRequest {
        id
        email
        name
        createdAt
        additionalInfo
        status
    }

    query GetAccessRequests($status: AccessRequestStatus!, $first: Int, $last: Int, $after: String, $before: String) {
        accessRequests(status: $status, first: $first, last: $last, after: $after, before: $before) {
            totalCount
            pageInfo {
                hasNextPage
                hasPreviousPage
                endCursor
                startCursor
            }
            nodes {
                ...AccessRequestNode
            }
        }
    }
`

/**
 * Graphql query to list the total number of registered users and the number license seats.
 */
export const HAS_LICENSE_SEATS = gql`
    query HasLicenseSeats {
        site {
            productSubscription {
                license {
                    tags
                    userCount
                }
            }
            users(deletedAt: { empty: true }) {
                totalCount
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
        accessRequestCreateUser(username: $username, email: $email, verifiedEmail: false) {
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
