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

const analyticsInsightStatItemFragment = gql`
    fragment AnalyticsInsightStatItemFragment on AnalyticsInsightStatItem {
        nodes {
            date
            count
        }
        summary {
            totalCount
        }
    }
`

export const INSIGHTS_STATISTICS = gql`
    query InsightsStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                codeInsights(dateRange: $dateRange, grouping: $grouping) {
                    dashboardCreations {
                        ...AnalyticsInsightStatItemFragment
                    }
                    insightHovers {
                        ...AnalyticsStatItemFragment
                    }
                    insightDataPointClicks {
                        ...AnalyticsStatItemFragment
                    }
                    totalInsightsCount
                }
            }
        }
    }
    ${analyticsStatItemFragment}
    ${analyticsInsightStatItemFragment}
`
