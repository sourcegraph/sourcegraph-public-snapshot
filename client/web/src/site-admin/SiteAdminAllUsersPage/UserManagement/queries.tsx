import { gql } from '@sourcegraph/http-client'

export const USERS_MANAGEMENT_SUMMARY = gql`
    query UsersManagementSummary {
        site {
            productSubscription {
                license {
                    userCount
                }
            }
            adminUsers: users(siteAdmin: true, deleted: false) {
                totalCount
            }
            users {
                totalCount
            }
        }
        registeredUsers: users {
            totalCount
        }
    }
`

export const USERS_MANAGEMENT_USERS_LIST = gql`
    query UsersManagementUsersList(
        $first: Int!
        $lastActivePeriod: SiteUsersLastActivePeriod
        $query: String
        $orderBy: SiteUserOrderBy
        $descending: Boolean
    ) {
        site {
            users(query: $query, lastActivePeriod: $lastActivePeriod) {
                totalCount
                nodes(first: $first, orderBy: $orderBy, descending: $descending) {
                    id
                    username
                    displayName
                    email
                    siteAdmin
                    eventsCount
                    createdAt
                    lastActiveAt
                    deletedAt
                }
            }
        }
    }
`

export const FORCE_SIGN_OUT_USERS = gql`
    mutation InvalidateSessionsByIDs($userIDs: [ID!]!) {
        invalidateSessionsByIDs(userIDs: $userIDs) {
            alwaysNil
        }
    }
`

export const DELETE_USERS = gql`
    mutation DeleteUsers($userIDs: [ID!]!) {
        deleteUsers(users: $userIDs) {
            alwaysNil
        }
    }
`

export const DELETE_USERS_FOREVER = gql`
    mutation DeleteUsersForever($userIDs: [ID!]!) {
        deleteUsers(users: $userIDs, hard: true) {
            alwaysNil
        }
    }
`
