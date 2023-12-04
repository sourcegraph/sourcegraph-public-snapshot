import { gql } from '@sourcegraph/http-client'

export const CodeInsightsJobFragment = gql`
    fragment InsightJob on InsightBackfillQueueItem {
        id
        insightViewTitle
        seriesLabel
        seriesSearchQuery
        backfillQueueStatus {
            state
            queuePosition
            errors
            cost
            percentComplete
            createdAt
            startedAt
            completedAt
            runtime
        }
    }
`

export const GET_CODE_INSIGHTS_JOBS = gql`
    query GetCodeInsightsJobs(
        $after: String
        $search: String
        $first: Int
        $last: Int
        $orderBy: BackfillQueueOrderBy
        $states: [InsightQueueItemState!]
    ) {
        insightAdminBackfillQueue(
            first: $first
            after: $after
            textSearch: $search
            orderBy: $orderBy
            states: $states
            last: $last
        ) {
            nodes {
                ...InsightJob
            }
            pageInfo {
                hasNextPage
                hasPreviousPage
                startCursor
                endCursor
            }
            totalCount
        }
    }
    ${CodeInsightsJobFragment}
`
