import { gql } from '@sourcegraph/http-client'

const analyticsStatItemFragment = gql`
    fragment SearchStatItemFragment on AnalyticsStatItem {
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

export const SEARCH_STATISTICS = gql`
    query SearchStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                search(dateRange: $dateRange, grouping: $grouping) {
                    searches {
                        ...SearchStatItemFragment
                    }
                    resultClicks {
                        ...SearchStatItemFragment
                    }
                    fileViews {
                        ...SearchStatItemFragment
                    }
                    fileOpens {
                        ...SearchStatItemFragment
                    }
                    codeCopied {
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
