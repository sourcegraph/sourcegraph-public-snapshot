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
                    seriesCreations {
                        ...AnalyticsInsightStatItemFragment
                    }
                    dashboardCreations {
                        ...AnalyticsInsightStatItemFragment
                    }
                    insightHovers {
                        ...AnalyticsStatItemFragment
                    }
                    insightDataPointClicks {
                        ...AnalyticsStatItemFragment
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
                    totalInsightsCount
                }
            }
        }
    }
    ${analyticsStatItemFragment}
    ${analyticsInsightStatItemFragment}
`
