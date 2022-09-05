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

export const CODEINTEL_STATISTICS = gql`
    query CodeIntelStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                repos {
                    count
                    preciseCodeIntelCount
                }
                codeIntel(dateRange: $dateRange, grouping: $grouping) {
                    referenceClicks {
                        ...AnalyticsStatItemFragment
                    }
                    definitionClicks {
                        ...AnalyticsStatItemFragment
                    }
                    inAppEvents {
                        summary {
                            totalCount
                        }
                    }
                    codeHostEvents {
                        summary {
                            totalCount
                        }
                    }
                    searchBasedEvents {
                        summary {
                            totalCount
                        }
                    }
                    preciseEvents {
                        summary {
                            totalCount
                        }
                    }
                    crossRepoEvents {
                        summary {
                            totalCount
                        }
                    }
                }
                codeIntelByLanguage(dateRange: $dateRange) {
                    language
                    precision
                    count
                }
                codeIntelTopRepositories(dateRange: $dateRange) {
                    name
                    language
                    kind
                    precision
                    events
                    hasPrecise
                }
            }
        }
    }
    ${analyticsStatItemFragment}
`
