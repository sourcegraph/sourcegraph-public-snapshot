import { Duration } from 'date-fns'
import { uniq } from 'lodash'

import { isDefined } from '@sourcegraph/shared/src/util/types'

import { InsightViewNode, TimeIntervalStepInput, TimeIntervalStepUnit } from '../../../../../graphql-operations'
import { Insight, InsightExecutionType, InsightType, SearchBasedInsight } from '../../types'

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

/**
 * Returns tuple with gql model time interval unit and value. Used to convert FE model
 * insight data series time step to GQL time interval model.
 */
export function getStepInterval(insight: SearchBasedInsight): [TimeIntervalStepUnit, number] {
    if (insight.type === InsightExecutionType.Backend) {
        return [TimeIntervalStepUnit.WEEK, 2]
    }

    const castUnits = (Object.keys(insight.step) as (keyof Duration)[])
        .map<[TimeIntervalStepUnit, number] | null>(key => {
            switch (key) {
                case 'hours':
                    return [TimeIntervalStepUnit.HOUR, insight.step[key] ?? 0]
                case 'days':
                    return [TimeIntervalStepUnit.DAY, insight.step[key] ?? 0]
                case 'weeks':
                    return [TimeIntervalStepUnit.WEEK, insight.step[key] ?? 0]
                case 'months':
                    return [TimeIntervalStepUnit.MONTH, insight.step[key] ?? 0]
                case 'years':
                    return [TimeIntervalStepUnit.YEAR, insight.step[key] ?? 0]
            }

            return null
        })
        .filter(isDefined)

    if (castUnits.length === 0) {
        throw new Error('Wrong time step format')
    }

    // Return first valid match
    return castUnits[0]
}

/**
 * Transforms/casts gql api insight model to FE insight model. We still
 * need to have specific FE model in order to support setting-based api
 * approach. This transformer will be removed as soon as we sunset setting-based
 * api for insights.
 *
 * @param insight - gql insight model
 */
export const getInsightView = (insight: InsightViewNode): Insight | undefined => {
    switch (insight.presentation.__typename) {
        case 'LineChartInsightViewPresentation': {
            const isBackendInsight = insight.dataSeriesDefinitions.every(
                series => series.repositoryScope.repositories.length === 0
            )

            const series = insight.presentation.seriesPresentation.map(series => ({
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

                    // In gql api we don't have this concept as visibility on FE.
                    // Insights have special system about visibility on BE only.
                    visibility: '',
                }
            }

            const repositories = uniq(
                insight.dataSeriesDefinitions.flatMap(series => series.repositoryScope.repositories)
            )

            const step = getDurationFromStep(insight.dataSeriesDefinitions[0].timeScope)

            return {
                id: insight.id,
                type: InsightExecutionType.Runtime,
                viewType: InsightType.SearchBased,
                title: insight.presentation.title,
                step,
                repositories,
                series,

                // In gql api we don't have this concept as visibility on FE.
                // Insights have special system about visibility on BE only.
                visibility: '',
            }
        }

        case 'PieChartInsightViewPresentation': {
            // At the moment we BE doesn't have special fragment type for Lang Stats repositories.
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

                // In gql api we don't have this concept as visibility on FE.
                // Insights have special system about visibility on BE only.
                visibility: '',
            }
        }
    }
}
