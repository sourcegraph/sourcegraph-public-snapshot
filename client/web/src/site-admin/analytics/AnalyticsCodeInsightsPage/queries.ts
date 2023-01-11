import { gql } from '@sourcegraph/http-client'

const analyticsStatItemFragment = gql`
    fragment InsightsStatItemFragment on AnalyticsStatItem {
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
`

export const INSIGHTS_STATISTICS = gql`
    query InsightsStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                codeInsights(dateRange: $dateRange, grouping: $grouping) {
                    insightHovers {
                        ...InsightsStatItemFragment
                    }
                    insightDataPointClicks {
                        ...InsightsStatItemFragment
                    }
                }
            }
        }
        insightViews {
            nodes {
                id
            }
        }
        insightsDashboards {
            nodes {
                id
            }
        }
    }
    ${analyticsStatItemFragment}
`
