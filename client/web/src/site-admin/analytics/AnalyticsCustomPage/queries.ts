import { gql } from '@sourcegraph/http-client'

export const CUSTOM_STATISTICS = gql`
    query CustomStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!, $events: [String!]!) {
        site {
            analytics {
                custom(dateRange: $dateRange, grouping: $grouping, events: $events) {
                    users {
                        nodes {
                            date
                            count
                            uniqueUsers
                            registeredUsers
                        }
                        summary {
                            totalCount
                            totalUniqueUsers
                            totalRegisteredUsers
                        }
                    }
                }
            }
        }
    }
`

export const CUSTOM_USERS_CONNECTION = gql`
    query CustomUsersConnection(
        $dateRange: AnalyticsDateRange!
        $grouping: AnalyticsGrouping!
        $events: [String!]!
        $first: Int
        $after: String
    ) {
        site {
            analytics {
                custom(dateRange: $dateRange, grouping: $grouping, events: $events) {
                    userActivity(first: $first, after: $after) {
                        nodes {
                            userID
                            username
                            displayName
                            totalEventCount
                            periods {
                                date
                                count
                            }
                        }
                        totalCount
                        pageInfo {
                            endCursor
                            hasNextPage
                        }
                    }
                }
            }
        }
    }
`

export const ALL_EVENT_NAMES = gql`
    query AllEventNames {
        site {
            analytics {
                allEventNames
            }
        }
    }
`
