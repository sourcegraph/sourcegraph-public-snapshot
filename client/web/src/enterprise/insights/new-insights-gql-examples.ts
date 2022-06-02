import { gql } from '@apollo/client';

/**
 * This query will run when we open the dashboard page
 * At mount of this page we don't need insight data because it will be loaded
 * later on demands (by lazy fetching or based on virtual scrolling parameters)
 *
 * In this query we load only vital information for insight card with loading state
 * all other fields/info will be loaded later.
 */
export const GET_INSIGHT_CONFIGURATIONS = gql`
    query GetInsightConfiguration {
        newInsights {
            nodes {
                id
                title
                isFrozen
                filters

                configuration {
                    ... on SearchBasedInsightConfiguration {
                        series {
                            id
                            title
                            visualSettings {
                                color
                            }
                        }
                    }

                   # All other insight types don't have important
                   # settings for the dashboard page mount
                }
            }
        }
    }
`

/**
 * This query is used when we want to load data for particular insight
 * on the dashboard or standalone insight pages. As you can see we do
 * not load anything about insight configuration but fetch insight data
 * (lines, categories, ...etc)
 */
export const GET_INSIGHT_DATA = gql`
    query GetInsightData($id: ID! $filters: NewInsightFilters) {
        newInsights(id: $id, filters: $filters) {
            nodes {
                id

                # Since we're carrying only about insight data here
                # we use data separation here SeriesLikeInsight and
                # CategoricalLikeInsight interfaces
                configuration {
                    ... on SeriesLikeInsight {
                        series {
                            id
                            title
                            status {
                                pendingJobs
                                backfillQueuedAt
                            }
                            points {
                                value
                                dateTime
                            }
                            visualSettings {
                                color
                                chartType
                                pattern
                            }
                        }
                    }

                    ... on CategoricalLikeInsight {
                        data {
                            value
                            title
                            color
                            link
                        }
                    }
                }
            }
        }
    }
`

/**
 * This query is used for the EDIT UI pages, it contains only insight
 * configuration and all top level base information that is required for
 * the edit UI.
 */
export const GET_INSIGHT_EDIT_INFORMATION = gql`
    query GetInsightEditInformation {
        newInsights {
            nodes {
                id
                title
                dashboards {
                    nodes {
                        id
                        title
                    }
                }

                # Here we use insight type separation because this is
                # how we separate insight in the product so it makes sense
                # to separate them in a similar way in API for consumers
                # As you can see there is no FE/consumer mapping here between
                # insight series and its configuration
                configuration {
                    ... on SearchBasedInsightConfiguration {
                        series {
                            id
                            title
                            query
                            repositoryScope {
                                repositories
                            }
                            timeScope
                            visualSettings {
                                color
                                chartType
                            }
                        }
                    }

                    ... on CaptureGroupInsightConfiguration {
                        query
                        repositories
                        timeScope
                    }

                    ... on LangStatsInsightConfiguration {
                        otherThreshold
                        repositories
                    }
                }
            }
        }
    }
`

// Mutation examples is kind of straightforward, so I assume that the
// examples of GQL schema would be enough to understand my proposal

