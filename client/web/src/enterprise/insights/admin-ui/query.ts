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
        $orderBy: BackfillQueueOrderBy
        $states: [InsightQueueItemState!]
    ) {
        insightAdminBackfillQueue(first: 10, after: $after, textSearch: $search, orderBy: $orderBy, states: $states) {
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
