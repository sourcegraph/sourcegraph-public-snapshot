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

export const NOTEBOOKS_STATISTICS = gql`
    query NotebooksStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                notebooks(dateRange: $dateRange, grouping: $grouping) {
                    creations {
                        ...AnalyticsStatItemFragment
                    }
                    views {
                        ...AnalyticsStatItemFragment
                    }
                    blockRuns {
                        summary {
                            totalCount
                            totalRegisteredUsers
                        }
                    }
                }
            }
        }
    }
    ${analyticsStatItemFragment}
`
