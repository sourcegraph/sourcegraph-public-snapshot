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
                    seriesCreations {
                        summary {
                            totalCount
                        }
                        nodes {
                            date
                            count
                        }
                    }
                    searchSeriesCreations: seriesCreations(generationType: SEARCH) {
                        summary {
                            totalCount
                        }
                    }
                    languageSeriesCreations: seriesCreations(generationType: LANGUAGE_STATS) {
                        summary {
                            totalCount
                        }
                    }
                    computeSeriesCreations: seriesCreations(generationType: SEARCH_COMPUTE) {
                        summary {
                            totalCount
                        }
                    }
                    insightHovers {
                        ...AnalyticsStatItemFragment
                    }
                    insightDataPointClicks {
                        ...AnalyticsStatItemFragment
                    }
                    summary {
                        totalInsightsCount
                        totalDashboardsCount
                    }
                }
            }
        }
    }
    ${analyticsStatItemFragment}
`
