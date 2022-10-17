import { gql } from '@sourcegraph/http-client'

const analyticsStatItemFragment = gql`
    fragment NotebooksStatItemFragment on AnalyticsStatItem {
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

export const NOTEBOOKS_STATISTICS = gql`
    query NotebooksStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                notebooks(dateRange: $dateRange, grouping: $grouping) {
                    creations {
                        ...NotebooksStatItemFragment
                    }
                    views {
                        ...NotebooksStatItemFragment
                    }
                    blockRuns {
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
    ${analyticsStatItemFragment}
`
