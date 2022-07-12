import { gql } from '@sourcegraph/http-client'

const analyticsStatItemFragment = gql`
    fragment AnalyticsStatItemFragment on AnalyticsStatItem {
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
`

const analyticsStatItemSummaryFragment = gql`
    fragment AnalyticsStatItemSummaryFragment on AnalyticsStatItemSummary {
        totalCount
        totalUniqueUsers
        totalRegisteredUsers
    }
`

export const USERS_STATISTICS = gql`
    query UsersStatistics($dateRange: AnalyticsDateRange!) {
        site {
            analytics {
                users(dateRange: $dateRange) {
                    summary {
                        avgDAU {
                            ...AnalyticsStatItemSummaryFragment
                        }
                        avgWAU {
                            ...AnalyticsStatItemSummaryFragment
                        }
                        avgMAU {
                            ...AnalyticsStatItemSummaryFragment
                        }
                    }
                    activity {
                        ...AnalyticsStatItemFragment
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
    ${analyticsStatItemSummaryFragment}
    ${analyticsStatItemFragment}
`
