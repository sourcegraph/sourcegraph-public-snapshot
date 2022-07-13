import { gql } from '@sourcegraph/http-client'

export const BATCHCHANGES_STATISTICS = gql`
    query BatchChangesStatistics($dateRange: AnalyticsDateRange!) {
        site {
            analytics {
                batchChanges(dateRange: $dateRange) {
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
