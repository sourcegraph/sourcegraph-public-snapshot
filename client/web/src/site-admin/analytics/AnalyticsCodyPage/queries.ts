import { gql } from '@sourcegraph/http-client'

const analyticsStatItemFragment = gql`
    fragment CodyStatItemFragment on AnalyticsStatItem {
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
export const CODY_STATISTICS = gql`
    query CodyStatistics($dateRange: AnalyticsDateRange!, $grouping: AnalyticsGrouping!) {
        site {
            analytics {
                cody(dateRange: $dateRange, grouping: $grouping) {
                    users {
                        ...CodyStatItemFragment
                    }
                    prompts {
                        ...CodyStatItemFragment
                    }
                    completionsAccepted {
                        ...CodyStatItemFragment
                    }
                    completionsSuggested {
                        ...CodyStatItemFragment
                    }
                }
            }
        }
    }
    ${analyticsStatItemFragment}
`
