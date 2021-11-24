import { gql } from '@apollo/client'

/**
 * GQL query for fetching insight data model with data series points and chart
 * information.
 */
export const GET_INSIGHT_VIEW_GQL = gql`
    query GetInsightView($id: ID, $filters: InsightViewFiltersInput) {
        insightViews(id: $id, filters: $filters) {
            nodes {
                id
                dataSeries {
                    seriesId
                    label
                    points {
                        dateTime
                        value
                    }
                    status {
                        backfillQueuedAt
                        completedJobs
                        pendingJobs
                        failedJobs
                    }
                }
            }
        }
    }
`
