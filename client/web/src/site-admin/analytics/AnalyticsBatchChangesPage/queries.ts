import { gql } from '@sourcegraph/http-client'

const analyticsStatItemFragment = gql`
    fragment AnalyticsStatItemFragment on AnalyticsStatItem {
        nodes {
            date
            count
        }
        summary {
            totalCount
        }
    }
`

export const BATCHCHANGES_STATISTICS = gql`
    query BatchChangesStatistics($dateRange: AnalyticsDateRange!) {
        site {
            analytics {
                batchChanges(dateRange: $dateRange) {
                    changesetsCreated {
                        ...AnalyticsStatItemFragment
                    }
                    changesetsMerged {
                        ...AnalyticsStatItemFragment
                    }
                }
            }
        }
    }
    ${analyticsStatItemFragment}
`
