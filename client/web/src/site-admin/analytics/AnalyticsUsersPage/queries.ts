import { gql } from '@sourcegraph/http-client'

export const USERS_STATISTICS = gql`
    query UsersStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                users(dateRange: $dateRange, grouping: $grouping) {
                    monthlyActiveUsers {
                        date
                        count
                    }
                    activity {
                        nodes {
                            date
                            count
                            uniqueUsers
                        }
                        summary {
                            totalCount
                            totalUniqueUsers
                            totalRegisteredUsers
                        }
                    }
                    frequencies {
                        daysUsed
                        frequency
                        percentage
                    }
                }
            }
            productSubscription {
                license {
                    userCount
                }
            }
        }
        users {
            totalCount
        }
        pendingAccessRequests: accessRequests(status: PENDING) {
            totalCount
        }
    }
`

export const TOTAL_USERS = gql`
    query UsersTotalCount {
        site {
            users {
                totalCount
            }
        }
    }
`

export const LICENSE_USER_LIMIT = gql`
    query LicenseUserLimit {
        site {
            productSubscription {
                license {
                    userCount
                }
            }
        }
    }
`
