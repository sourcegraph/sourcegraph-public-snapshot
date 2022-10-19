import { gql } from '@apollo/client'

/**
 * GQL query for fetching insight data model with data series points and chart
 * information.
 */
export const GET_INSIGHT_DATA = gql`
    query GetInsightData($id: ID, $filters: InsightViewFiltersInput, $seriesDisplayOptions: SeriesDisplayOptionsInput) {
        insightViews(id: $id, filters: $filters, seriesDisplayOptions: $seriesDisplayOptions) {
            nodes {
                ...InsightDataNode
            }
        }
    }

    fragment InsightDataNode on InsightView {
        id
        dataSeries {
            ...InsightDataSeries
        }
        #        status {
        #            progress
        #            errors {
        #                reason
        #            }
        #        }
    }

    fragment InsightDataSeries on InsightsSeries {
        seriesId
        label
        points {
            dateTime
            value
        }
        status {
            ...InsightDataSeriesStatus
        }
        #        status2 {
        #            ...InsightDataSeriesStatus2
        #        }
    }

    fragment InsightDataSeriesStatus on InsightSeriesStatus {
        backfillQueuedAt
        completedJobs
        pendingJobs
        failedJobs
    }

    #    fragment InsightDataSeriesStatus2 on InsightSeriesProgress {
    #        progress
    #        errors {
    #            reason
    #        }
    #    }
`
