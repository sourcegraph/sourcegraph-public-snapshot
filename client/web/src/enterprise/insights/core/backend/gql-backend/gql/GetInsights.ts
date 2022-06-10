import { gql } from '@apollo/client'

const INSIGHT_VIEW_SERIES_FRAGMENT = gql`
    fragment InsightViewSeries on InsightView {
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
                isCalculated
                generatedFromCaptureGroups
            }
        }
    }
`

export const INSIGHT_VIEW_FRAGMENT = gql`
    fragment InsightViewNode on InsightView {
        id
        defaultSeriesDisplayOptions {
            limit
            sortOptions {
                mode
                direction
            }
        }
        appliedSeriesDisplayOptions {
            limit
            sortOptions {
                mode
                direction
            }
        }
        isFrozen
        appliedFilters {
            includeRepoRegex
            excludeRepoRegex
            searchContexts
        }
        dashboardReferenceCount
        dashboards {
            nodes {
                id
                title
            }
        }
        ...InsightViewSeries
    }

    ${INSIGHT_VIEW_SERIES_FRAGMENT}
`

/**
 * GQL query for getting all insight views that are accessible for a user.
 * Note that this query doesn't contain any chart data series points.
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
