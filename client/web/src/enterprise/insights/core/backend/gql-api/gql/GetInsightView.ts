import { gql } from '@apollo/client'

const INSIGHT_DATA_SERIES_FRAGMENT = gql`
    fragment InsightDataSeries on InsightsSeries {
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
`

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
                    ...InsightDataSeries
                }
            }
        }
    }
    ${INSIGHT_DATA_SERIES_FRAGMENT}
`
