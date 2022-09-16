import { gql } from '@sourcegraph/http-client'

const analyticsStatItemFragment = gql`
    fragment AnalyticsStatItemFragment on AnalyticsStatItem {
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
`

export const SEARCH_STATISTICS = gql`
    query SearchStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                search(dateRange: $dateRange, grouping: $grouping) {
                    searches {
                        ...AnalyticsStatItemFragment
                    }
                    resultClicks {
                        ...AnalyticsStatItemFragment
                    }
                    fileViews {
                        ...AnalyticsStatItemFragment
                    }
                    fileOpens {
                        ...AnalyticsStatItemFragment
                    }
                }
            }
        }
    }
    ${analyticsStatItemFragment}
`
