import { gql } from '@sourcegraph/http-client'

export const USERS_STATISTICS = gql`
    query UsersStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                users(dateRange: $dateRange, grouping: $grouping) {
                    summary {
                        avgDAU
                        avgWAU
                        avgMAU
                    }
                    activity {
                        nodes {
                            date
                            count
                            registeredUsers
                        }
                        summary {
                            totalCount
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
    }
`
