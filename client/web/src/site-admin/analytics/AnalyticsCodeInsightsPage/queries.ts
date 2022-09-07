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

export const INSIGHTS_STATISTICS = gql`
    query InsightsStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                codeInsights(dateRange: $dateRange, grouping: $grouping) {
                    insightCreations {
                        ...AnalyticsStatItemFragment
                    }
                    # searchInsightCreations: insightCreations(generationType: SEARCH) {
                    #     ...AnalyticsStatItemFragment
                    # }
                    # languageInsightCreations: insightCreations(generationType: LANGUAGE_STATS) {
                    #     ...AnalyticsStatItemFragment
                    # }
                    # computeInsightCreations: insightCreations(generationType: SEARCH_COMPUTE) {
                    #     ...AnalyticsStatItemFragment
                    # }
                    insightHovers {
                        ...AnalyticsStatItemFragment
                    }
                    insightDataPointClicks {
                        ...AnalyticsStatItemFragment
                    }
                    # dashboardCreations {
                    #     ...AnalyticsStatItemFragment
                    # }
                }
            }
        }
    }
    ${analyticsStatItemFragment}
`
