import { Duration } from 'date-fns'
import { uniq } from 'lodash'

import { InsightViewNode, TimeIntervalStepInput, TimeIntervalStepUnit } from '../../../../../../graphql-operations'
import { Insight, InsightExecutionType, InsightType } from '../../../types'

/**
 * Transforms/casts gql api insight model to FE insight model. We still
 * need to have a specific FE model in order to support setting-based api
 * approach. This transformer will be removed as soon as we sunset setting-based
 * api for insights.
 */
export const createInsightView = (insight: InsightViewNode): Insight => {
    switch (insight.presentation.__typename) {
        case 'LineChartInsightViewPresentation': {
            const isBackendInsight = insight.dataSeriesDefinitions.every(series => series.isCalculated)
            const isCaptureGroupInsight = insight.dataSeriesDefinitions.some(
                series => series.generatedFromCaptureGroups
            )
            // We do not support different time scope for different series at the moment
            const step = getDurationFromStep(insight.dataSeriesDefinitions[0].timeScope)
            const repositories = uniq(
                insight.dataSeriesDefinitions.flatMap(series => series.repositoryScope.repositories)
            )

            if (isCaptureGroupInsight) {
                // It's safe because capture group insight always has only 1 data series
                const { query } = insight.dataSeriesDefinitions[0] ?? {}

                return {
                    id: insight.id,
                    type: InsightExecutionType.Backend,
                    viewType: InsightType.CaptureGroup,
                    title: insight.presentation.title,
                    repositories,
                    query,
                    step,
                    filters: {
                        includeRepoRegexp: insight.appliedFilters.includeRepoRegex ?? '',
                        excludeRepoRegexp: insight.appliedFilters.excludeRepoRegex ?? '',
                    },
                    dashboardReferenceCount: insight.dashboardReferenceCount,
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

            if (isBackendInsight) {
                return {
                    id: insight.id,
                    type: InsightExecutionType.Backend,
                    viewType: InsightType.SearchBased,
                    title: insight.presentation.title,
                    series,
                    step,
                    filters: {
                        includeRepoRegexp: insight.appliedFilters.includeRepoRegex ?? '',
                        excludeRepoRegexp: insight.appliedFilters.excludeRepoRegex ?? '',
                    },
                    dashboardReferenceCount: insight.dashboardReferenceCount,
                }
            }

            return {
                id: insight.id,
                type: InsightExecutionType.Runtime,
                viewType: InsightType.SearchBased,
                title: insight.presentation.title,
                step,
                repositories,
                series,
                dashboardReferenceCount: insight.dashboardReferenceCount,
            }
        }

        case 'PieChartInsightViewPresentation': {
            // At the moment BE doesn't have a special fragment type for Lang Stats repositories.
            // We use search based definition (first repo of first definition). For lang-stats
            // it always should be exactly one series with repository scope info.
            const repository = insight.dataSeriesDefinitions[0].repositoryScope.repositories[0] ?? ''

            return {
                id: insight.id,
                type: InsightExecutionType.Runtime,
                viewType: InsightType.LangStats,
                title: insight.presentation.title,
                otherThreshold: insight.presentation.otherThreshold,
                repository,
                dashboardReferenceCount: insight.dashboardReferenceCount,
            }
        }
    }
}

function getDurationFromStep(step: TimeIntervalStepInput): Duration {
    switch (step.unit) {
        case TimeIntervalStepUnit.HOUR:
            return { hours: step.value }
        case TimeIntervalStepUnit.DAY:
            return { days: step.value }
        case TimeIntervalStepUnit.WEEK:
            return { weeks: step.value }
        case TimeIntervalStepUnit.MONTH:
            return { months: step.value }
        case TimeIntervalStepUnit.YEAR:
            return { years: step.value }
    }
}
