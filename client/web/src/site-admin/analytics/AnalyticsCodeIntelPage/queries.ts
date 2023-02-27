import { gql } from '@sourcegraph/http-client'

const analyticsStatItemFragment = gql`
    fragment CodeIntelStatItemFragment on AnalyticsStatItem {
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
                        ...CodeIntelStatItemFragment
                    }
                    definitionClicks {
                        ...CodeIntelStatItemFragment
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
