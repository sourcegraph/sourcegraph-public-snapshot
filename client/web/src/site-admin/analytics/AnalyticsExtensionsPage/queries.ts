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

export const EXTENSIONS_STATISTICS = gql`
    query ExtensionsStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                extensions(dateRange: $dateRange, grouping: $grouping) {
                    jetbrains {
                        ...AnalyticsStatItemFragment
                    }
                    vscode {
                        ...AnalyticsStatItemFragment
                    }
                    browser {
                        ...AnalyticsStatItemFragment
                    }
                }
            }
            users(deletedAt: { empty: true }) {
                totalCount
            }
        }
    }
    ${analyticsStatItemFragment}
`
