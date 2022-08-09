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

export const FORCE_SIGN_OUT_USER = gql`
    mutation InvalidateSessionsByIDs {
        invalidateSessionsByIDs(userIDs: ["userID1", "userID2"]) {
            alwaysNil
        }
    }
`
