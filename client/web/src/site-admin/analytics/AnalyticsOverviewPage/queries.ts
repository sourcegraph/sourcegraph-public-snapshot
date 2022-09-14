import { gql } from '@sourcegraph/http-client'

export const OVERVIEW_STATISTICS = gql`
    query OverviewStatistics {
        site {
            productVersion
            productSubscription {
                productNameWithBrand
                actualUserCount
                license {
                    userCount
                    expiresAt
                }
            }
            adminUsers: users(siteAdmin: true, deletedAt: { empty: true }) {
                totalCount
            }
        }
        users {
            totalCount
        }
        repositories {
            totalCount(precise: true)
        }
        repositoryStats {
            gitDirBytes
            indexedLinesCount
        }
        surveyResponses {
            totalCount
            averageScore
            netPromoterScore
        }
    }
`

export const OVERVIEW_DEV_TIME_SAVED = gql`
    query OverviewDevTimeSaved($dateRange: AnalyticsDateRange!) {
        site {
            analytics {
                search(dateRange: $dateRange, grouping: WEEKLY) {
                    searches {
                        summary {
                            totalCount
                        }
                    }
                    fileViews {
                        summary {
                            totalCount
                        }
                    }
                }
                codeIntel(dateRange: $dateRange, grouping: WEEKLY) {
                    referenceClicks {
                        summary {
                            totalCount
                        }
                    }
                    definitionClicks {
                        summary {
                            totalCount
                        }
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
                batchChanges(dateRange: $dateRange, grouping: WEEKLY) {
                    changesetsMerged {
                        summary {
                            totalCount
                        }
                    }
                }
                notebooks(dateRange: $dateRange, grouping: WEEKLY) {
                    views {
                        summary {
                            totalCount
                        }
                    }
                }
                extensions(dateRange: $dateRange, grouping: WEEKLY) {
                    jetbrains {
                        summary {
                            totalCount
                        }
                    }
                    vscode {
                        summary {
                            totalCount
                        }
                    }
                    browser {
                        summary {
                            totalCount
                        }
                    }
                }
                users(dateRange: $dateRange, grouping: WEEKLY) {
                    activity {
                        summary {
                            totalRegisteredUsers
                        }
                    }
                }
            }
        }
    }
`
