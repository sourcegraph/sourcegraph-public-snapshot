import { type InsightViewNode, GroupByField } from '../../../../../../graphql-operations'
import { type ComputeInsight, type Insight, InsightType } from '../../../types'
import type { BaseInsight } from '../../../types/insight/common'

import { getDurationFromStep, getInsightRepositories, getParsedFilters } from './field-parsers'

/**
 * Transforms/casts gql api insight model to FE insight model. We still
 * need to have a specific FE model in order to support setting-based api
 * approach. This transformer will be removed as soon as we sunset setting-based
 * api for insights.
 */
export const createInsightView = (insight: InsightViewNode): Insight => {
    const baseInsight: Omit<BaseInsight, 'type'> = {
        id: insight.id,
        title: insight.presentation.title,
        isFrozen: insight.isFrozen,
        dashboards: insight.dashboards?.nodes ?? [],
        dashboardReferenceCount: insight.dashboardReferenceCount,
    }

    switch (insight.presentation.__typename) {
        case 'LineChartInsightViewPresentation': {
            const isComputeInsight = insight.dataSeriesDefinitions.some(series => series.groupBy)
            const isCaptureGroupInsight = insight.dataSeriesDefinitions.some(
                series => series.generatedFromCaptureGroups && !series.groupBy
            )

            const { defaultFilters, defaultSeriesDisplayOptions } = insight
            // We do not support different time scope for different series at the moment
            const step = getDurationFromStep(insight.dataSeriesDefinitions[0].timeScope)
            const { repositories, repoSearch } = getInsightRepositories(insight.repositoryDefinition)
            const filters = getParsedFilters(defaultFilters, defaultSeriesDisplayOptions)

            if (isCaptureGroupInsight) {
                // It's safe because capture group insight always has only 1 data series
                const { query } = insight.dataSeriesDefinitions[0] ?? {}

                return {
                    ...baseInsight,
                    type: InsightType.CaptureGroup,
                    repositories,
                    repoQuery: repoSearch,
                    query,
                    step,
                    filters,
                }
            }

            const series = insight.presentation.seriesPresentation.map(series => ({
                id: series.seriesId,
                name: series.label,
                query:
                    insight.dataSeriesDefinitions.find(definition => definition.seriesId === series.seriesId)?.query ||
                    'QUERY NOT FOUND',
                stroke:
                    'seriesPresentation' in insight.presentation
                        ? insight.presentation.seriesPresentation.find(
                              presentation => presentation.seriesId === series.seriesId
                          )?.color
                        : '',
            }))

            if (isComputeInsight) {
                // It's safe because capture group insight always has only 1 data series
                const { groupBy } = insight.dataSeriesDefinitions[0] ?? {}

                return {
                    ...baseInsight,
                    type: InsightType.Compute,
                    groupBy: groupBy ?? GroupByField.REPO,
                    repositories,
                    series,
                    filters,
                } as ComputeInsight
            }

            return {
                ...baseInsight,
                type: InsightType.SearchBased,
                repositories,
                repoQuery: repoSearch,
                series,
                step,
                filters,
            }
        }

        case 'PieChartInsightViewPresentation': {
            // At the moment BE doesn't have a special fragment type for Lang Stats repositories.
            // We use search based definition (first repo of first definition). For lang-stats
            // it always should be exactly one series with repository scope info.
            const { repositories } = getInsightRepositories(insight.repositoryDefinition)

            return {
                ...baseInsight,
                type: InsightType.LangStats,
                title: insight.presentation.title,
                otherThreshold: insight.presentation.otherThreshold,
                repository: repositories[0] ?? '',
            }
        }
    }
}
