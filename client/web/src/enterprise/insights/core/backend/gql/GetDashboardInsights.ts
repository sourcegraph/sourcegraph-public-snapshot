import { gql } from '@apollo/client'

export const GET_DASHBOARD_INSIGHTS_GQL = gql`
    query GetDashboardInsights($id: ID) {
        insightsDashboards(id: $id) {
            nodes {
                views {
                    nodes {
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
                }
            }
        }
    }
`
