import { gql } from '@apollo/client'

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

    fragment InsightDataNode on InsightView {
        id
        dataSeries {
            ...InsightDataSeries
        }
    }

    fragment InsightDataSeries on InsightsSeries {
        seriesId
        label
        points {
            dateTime
            value
            pointInTimeQuery
        }
        status {
            isLoadingData
            incompleteDatapoints {
                ... on TimeoutDatapointAlert {
                    __typename
                    time
                }
                ... on GenericIncompleteDatapointAlert {
                    __typename
                    time
                    reason
                }
            }
        }
    }
`
