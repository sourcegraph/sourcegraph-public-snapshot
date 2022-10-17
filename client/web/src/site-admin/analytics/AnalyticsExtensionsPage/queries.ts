import { gql } from '@sourcegraph/http-client'

const analyticsStatItemFragment = gql`
    fragment ExtensionsStatItemFragment on AnalyticsStatItem {
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

export const EXTENSIONS_STATISTICS = gql`
    query ExtensionsStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                extensions(dateRange: $dateRange, grouping: $grouping) {
                    jetbrains {
                        ...ExtensionsStatItemFragment
                    }
                    vscode {
                        ...ExtensionsStatItemFragment
                    }
                    browser {
                        ...ExtensionsStatItemFragment
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
