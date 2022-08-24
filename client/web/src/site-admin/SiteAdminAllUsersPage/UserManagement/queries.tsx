import { gql } from '@sourcegraph/http-client'

export const USERS_MANAGEMENT_SUMMARY = gql`
    query UsersManagementSummary {
        site {
            productSubscription {
                license {
                    userCount
                }
            }
            adminUsers: users(siteAdmin: true, deletedAt: { isNull: true }) {
                totalCount
            }
            registeredUsers: users(deletedAt: { isNull: true }) {
                totalCount
            }
        }
    }
`

export const USERS_MANAGEMENT_USERS_LIST = gql`
    query UsersManagementUsersList(
        $first: Int!
        $lastActiveAt: SiteUsersNullableDateRangeInput
        $deletedAt: SiteUsersNullableDateRangeInput
        $createdAt: SiteUsersDateRangeInput
        $eventsCount: SiteUsersNumberRangeInput
        $query: String
        $orderBy: SiteUserOrderBy
        $descending: Boolean
        $siteAdmin: Boolean
    ) {
        site {
            users(
                query: $query
                lastActiveAt: $lastActiveAt
                siteAdmin: $siteAdmin
                deletedAt: $deletedAt
                createdAt: $createdAt
                eventsCount: $eventsCount
            ) {
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
