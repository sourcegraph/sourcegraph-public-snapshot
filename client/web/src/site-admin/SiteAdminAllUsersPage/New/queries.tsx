import { gql } from '@sourcegraph/http-client'

export const USERS_MANAGEMENT = gql`
    query UsersManagement(
        $dateRange: AnalyticsDateRange!
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
                nodes(first: 100, orderBy: $usersOrderBy, descending: $usersOrderDescending) {
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
