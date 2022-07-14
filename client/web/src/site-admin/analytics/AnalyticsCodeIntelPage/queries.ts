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
        users {
            totalCount
        }
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
