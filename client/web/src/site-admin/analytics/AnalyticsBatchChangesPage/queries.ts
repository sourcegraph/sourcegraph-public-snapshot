import { gql } from '@sourcegraph/http-client'

export const BATCHCHANGES_STATISTICS = gql`
    query BatchChangesStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                batchChanges(dateRange: $dateRange, grouping: $grouping) {
                    changesetsCreated {
                        nodes {
                            date
                            count
                        }
                        summary {
                            totalCount
                        }
                    }
                    changesetsMerged {
                        nodes {
                            date
                            count
                        }
                        summary {
                            totalCount
                        }
                    }
                }
            }
        }
    }
`
