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

const analyticsStatItemSummaryFragment = gql`
    fragment AnalyticsStatItemSummaryFragment on AnalyticsStatItemSummary {
        totalCount
        totalUniqueUsers
        totalRegisteredUsers
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
                    resultClicks {
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

export const USERS_STATISTICS = gql`
    query UsersStatistics($dateRange: AnalyticsDateRange!) {
        site {
            analytics {
                users(dateRange: $dateRange) {
                    summary {
                        avgDAU {
                            ...AnalyticsStatItemSummaryFragment
                        }
                        avgWAU {
                            ...AnalyticsStatItemSummaryFragment
                        }
                        avgMAU {
                            ...AnalyticsStatItemSummaryFragment
                        }
                    }
                    activity {
                        ...AnalyticsStatItemFragment
                    }
                    frequencies {
                        daysUsed
                        frequency
                        percentage
                    }
                }
            }
            productSubscription {
                license {
                    userCount
                }
            }
        }
        users {
            totalCount
        }
    }
    ${analyticsStatItemSummaryFragment}
    ${analyticsStatItemFragment}
`

export const CODEINTEL_STATISTICS = gql`
    query CodeIntelStatistics($dateRange: AnalyticsDateRange!) {
        currentUser {
            organizationMemberships {
                totalCount
            }
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
