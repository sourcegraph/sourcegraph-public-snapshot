import { gql } from '@sourcegraph/http-client'

export const USERS_MANAGEMENT = gql`
    query UsersManagement(
        $dateRange: AnalyticsDateRange!
        $first: Int!
        $grouping: AnalyticsGrouping!
        $usersLastActivePeriod: SiteUsersLastActivePeriod
        $usersQuery: String
        $usersOrderBy: SiteUserOrderBy
        $usersOrderDescending: Boolean
    ) {
        site {
            analytics {
                users(dateRange: $dateRange, grouping: $grouping) {
                    activity {
                        nodes {
                            date
                            count
                            uniqueUsers
                        }
                        summary {
                            totalCount
                            totalUniqueUsers
                        }
                    }
                }
            }
            productSubscription {
                license {
                    userCount
                }
            }
            users(query: $usersQuery, lastActivePeriod: $usersLastActivePeriod) {
                totalCount
                nodes(first: $first, orderBy: $usersOrderBy, descending: $usersOrderDescending) {
                    id
                    username
                    email
                    siteAdmin
                    eventsCount
                    createdAt
                    lastActiveAt
                    deletedAt
                }
            }
            adminUsers: users(siteAdmin: true) {
                totalCount
            }
        }
        users {
            totalCount
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

// TODO: reset password
// TODO: revoke/promote to site admin
