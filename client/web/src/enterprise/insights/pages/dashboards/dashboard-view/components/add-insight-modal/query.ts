import type { ApolloClient, ApolloError } from '@apollo/client'
import { groupBy } from 'lodash'

import { isDefined } from '@sourcegraph/common'
import { dataOrThrowErrors, getDocumentNode, gql } from '@sourcegraph/http-client'

import type { Connection } from '../../../../../../../components/FilteredConnection'
import { useShowMorePagination } from '../../../../../../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    type AssignableInsight,
    type DashboardInsights,
    type FindInsightsBySearchTermResult,
    type FindInsightsBySearchTermVariables,
    GroupByField,
} from '../../../../../../../graphql-operations'

import { type DashboardInsight, type InsightSuggestion, InsightType } from './types'

const SYNC_DASHBOARD_INSIGHTS = gql`
    fragment DashboardInsights on InsightsDashboard {
        views {
            nodes {
                id
                presentation {
                    __typename
                    ... on LineChartInsightViewPresentation {
                        title
                    }
                    ... on PieChartInsightViewPresentation {
                        title
                    }
                }
            }
        }
    }
`

/**
 * This cache function returns a minimal data object from the apollo cache about
 * insights that the currently viewed dashboard has see {@link GET_DASHBOARD_INSIGHTS_GQL}
 */
export function getCachedDashboardInsights(client: ApolloClient<unknown>, dashboardId: string): DashboardInsight[] {
    const dashboardInsights =
        client
            .readFragment<DashboardInsights>({
                id: `InsightsDashboard:${dashboardId}`,
                fragment: getDocumentNode(SYNC_DASHBOARD_INSIGHTS),
            })
            ?.views?.nodes.filter(isDefined) ?? []

    return dashboardInsights.map(insight => ({ id: insight.id, title: insight.presentation.title }))
}

export const GET_INSIGHTS_BY_SEARCH_TERM = gql`
    query FindInsightsBySearchTerm($search: String!, $first: Int, $after: String, $excludeIds: [ID!]) {
        insightViews(find: $search, first: $first, after: $after, excludeIds: $excludeIds) {
            nodes {
                ...AssignableInsight
            }
            totalCount
            pageInfo {
                hasNextPage
                endCursor
            }
        }
    }

    fragment AssignableInsight on InsightView {
        id
        presentation {
            __typename
            ... on LineChartInsightViewPresentation {
                title
            }
            ... on PieChartInsightViewPresentation {
                title
            }
        }
        dataSeriesDefinitions {
            ... on SearchInsightDataSeriesDefinition {
                query
                groupBy
                generatedFromCaptureGroups
            }
        }
    }
`

interface UseInsightSuggestionsInput {
    search: string
    excludeIds: string[]
}

interface UseInsightSuggestionsResult {
    connection?: Connection<InsightSuggestion>
    loading: boolean
    hasNextPage: boolean
    error?: ApolloError
    fetchMore: () => void
}

export function useInsightSuggestions(input: UseInsightSuggestionsInput): UseInsightSuggestionsResult {
    const { search, excludeIds } = input

    const { connection, loading, hasNextPage, error, fetchMore } = useShowMorePagination<
        FindInsightsBySearchTermResult,
        FindInsightsBySearchTermVariables,
        AssignableInsight | null
    >({
        query: GET_INSIGHTS_BY_SEARCH_TERM,
        variables: { first: 20, after: null, search, excludeIds },
        getConnection: result => {
            const { insightViews } = dataOrThrowErrors(result)

            return insightViews
        },
        options: { fetchPolicy: 'cache-and-network' },
    })

    if (!connection) {
        return {
            loading,
            hasNextPage,
            error,
            fetchMore,
        }
    }

    const normalizedConnection: Connection<InsightSuggestion> = {
        ...connection,
        nodes: makeInsightTitlesUnique(
            connection.nodes.filter<AssignableInsight>(
                (insight): insight is AssignableInsight => isDefined(insight) && !excludeIds.includes(insight.id)
            )
        ).map<InsightSuggestion>(insight => {
            const isLangStat = insight.presentation.__typename === 'PieChartInsightViewPresentation'

            if (isLangStat) {
                return {
                    id: insight.id,
                    title: insight.presentation.title,
                    type: InsightType.LanguageStats,
                }
            }

            const isCompute = insight.dataSeriesDefinitions.some(series => series.groupBy)

            if (isCompute) {
                const { groupBy, query } = insight.dataSeriesDefinitions[0] ?? {}
                return {
                    id: insight.id,
                    title: insight.presentation.title,
                    type: InsightType.Compute,
                    query,
                    groupBy: groupBy ?? GroupByField.AUTHOR,
                }
            }

            const isCaptureGroup = insight.dataSeriesDefinitions.some(
                series => series.generatedFromCaptureGroups && !series.groupBy
            )

            if (isCaptureGroup) {
                const { query } = insight.dataSeriesDefinitions[0] ?? {}

                return {
                    id: insight.id,
                    title: insight.presentation.title,
                    type: InsightType.DetectAndTrack,
                    query,
                }
            }

            return {
                id: insight.id,
                title: insight.presentation.title,
                type: InsightType.Detect,
                queries: insight.dataSeriesDefinitions.map(def => def.query),
            }
        }),
    }

    return {
        connection: normalizedConnection,
        loading,
        hasNextPage,
        error,
        fetchMore,
    }
}

function makeInsightTitlesUnique(insights: AssignableInsight[]): AssignableInsight[] {
    const groupedByTitle = groupBy(insights, insight => insight.presentation.title)

    return Object.keys(groupedByTitle).flatMap<AssignableInsight>(title => {
        if (groupedByTitle[title].length === 1) {
            return groupedByTitle[title]
        }

        return groupedByTitle[title].map((insight, index) => ({
            ...insight,
            presentation: {
                ...insight.presentation,
                title: `${insight.presentation.title} (${index + 1})`,
            },
        }))
    })
}
