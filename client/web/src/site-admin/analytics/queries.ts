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

export const SEARCH_STATISTICS = gql`
    query SearchStatistics($dateRange: AnalyticsDateRange!) {
        site {
            analytics {
                search(dateRange: $dateRange) {
                    searches {
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

export const NOTEBOOKS_STATISTICS = gql`
    query NotebooksStatistics($dateRange: AnalyticsDateRange!) {
        site {
            analytics {
                notebooks(dateRange: $dateRange) {
                    creations {
                        ...AnalyticsStatItemFragment
                    }
                    views {
                        ...AnalyticsStatItemFragment
                    }
                    blockRuns {
                        summary {
                            totalCount
                            totalUniqueUsers
                        }
                    }
                }
            }
        }
    }
    ${analyticsStatItemFragment}
`
