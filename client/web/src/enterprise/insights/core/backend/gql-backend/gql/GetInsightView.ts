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
            isLoadingData
        }
    }
`

const INSIGHT_DATA_NODE_FRAGMENT = gql`
    fragment InsightDataNode on InsightView {
        id
        dataSeries {
            ...InsightDataSeries
        }
    }
    ${INSIGHT_DATA_SERIES_FRAGMENT}
`

/**
 * GQL query for fetching insight data model with data series points and chart
 * information.
 */
export const GET_INSIGHT_VIEW_GQL = gql`
    query GetInsightView($id: ID, $filters: InsightViewFiltersInput, $seriesDisplayOptions: SeriesDisplayOptionsInput) {
        insightViews(id: $id, filters: $filters, seriesDisplayOptions: $seriesDisplayOptions) {
            nodes {
                ...InsightDataNode
            }
        }
    }
    ${INSIGHT_DATA_NODE_FRAGMENT}
`
