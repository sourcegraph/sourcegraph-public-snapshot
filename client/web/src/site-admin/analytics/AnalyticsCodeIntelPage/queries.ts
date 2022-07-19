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

export const CODEINTEL_STATISTICS = gql`
    query CodeIntelStatistics($dateRange: AnalyticsDateRange!) {
        site {
            analytics {
                repos {
                    count
                    preciseCodeIntelCount
                }
                codeIntel(dateRange: $dateRange) {
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
                    browserExtensionInstalls {
                        summary {
                            totalRegisteredUsers
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
            }
        }
    }
    ${analyticsStatItemFragment}
`
