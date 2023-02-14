import { gql } from '@sourcegraph/http-client'

export const PENDING_ACCESS_REQUESTS_LIST = gql`
    query PendingAccessRequestsList($limit: Int!, $offset: Int!) {
        accessRequests(status: PENDING) {
            totalCount
            nodes(limit: $limit, offset: $offset, orderBy: CREATED_AT, descending: true) {
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

export const REJECT_ACCESS_REQUEST = gql`
    mutation RejectAccessRequest($id: ID!) {
        rejectAccessRequest(id: $id) {
            alwaysNil
        }
    }
`
