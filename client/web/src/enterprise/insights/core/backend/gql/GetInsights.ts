import { gql } from '@apollo/client'

export const INSIGHT_VIEW_FRAGMENT = gql`
    fragment InsightViewNode on InsightView {
        id
        presentation {
            __typename
            ... on LineChartInsightViewPresentation {
                title
                seriesPresentation {
                    seriesId
                    label
                    color
                }
            }
            ... on PieChartInsightViewPresentation {
                title
                otherThreshold
            }
        }
        dataSeriesDefinitions {
            ... on SearchInsightDataSeriesDefinition {
                seriesId
                query
                repositoryScope {
                    repositories
                }
                timeScope {
                    ... on InsightIntervalTimeScope {
                        unit
                        value
                    }
                }
            }
        }
    }
`

/**
 * GQL query for getting all insight views that are accessible for a user.
 * Note that this query doesn't contains any chart data series points.
 * Insight model in this case contains only meta and presentation chart data.
 */
export const GET_INSIGHTS_GQL = gql`
    query GetInsights($id: ID) {
        insightViews(id: $id) {
            nodes {
                ...InsightViewNode
            }
        }
    }
    ${INSIGHT_VIEW_FRAGMENT}
`
