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
                timeScope {
                    ... on InsightIntervalTimeScope {
                        unit
                        value
                    }
                }
                isCalculated
                generatedFromCaptureGroups
                groupBy
            }
        }
    }
`

export const INSIGHT_VIEW_FRAGMENT = gql`
    fragment InsightViewNode on InsightView {
        id
        repositoryDefinition {
            __typename
            ... on RepositorySearchScope {
                search
            }
            ... on InsightRepositoryScope {
                repositories
            }
        }
        defaultSeriesDisplayOptions {
            limit
            numSamples
            sortOptions {
                mode
                direction
            }
        }
        isFrozen
        defaultFilters {
            includeRepoRegex
            excludeRepoRegex
            searchContexts
        }
        dashboardReferenceCount
        repositoryDefinition {
            __typename
            ... on RepositorySearchScope {
                search
            }

            ... on InsightRepositoryScope {
                repositories
            }
        }
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
