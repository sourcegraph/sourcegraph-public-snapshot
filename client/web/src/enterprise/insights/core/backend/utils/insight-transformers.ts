import { Duration } from 'date-fns'
import { uniq } from 'lodash'

import { isDefined } from '@sourcegraph/shared/src/util/types'

import { GetInsightsResult, InsightViewsFields, TimeIntervalStepUnit } from '../../../../../graphql-operations'
import { Insight, InsightType, SearchBasedInsight } from '../../types'

function getDurationFromStep(step: InsightViewsFields['dataSeriesDefinitions'][number]['timeScope']): Duration {
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
    if (insight.type === InsightType.Backend) {
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
export const getInsightView = (insight: GetInsightsResult['insightViews']['nodes'][0]): Insight => {
    // TODO [VK] Support lang stats insight
    switch (insight.presentation.__typename) {
        case 'LineChartInsightViewPresentation': {
            const isBackendInsight = insight.dataSeriesDefinitions.every(
                series => series.repositoryScope.repositories.length > 0
            )

            const series = insight.dataSeries.map(series => ({
                name: series.label,
                query:
                    insight.dataSeriesDefinitions.find(definition => definition.seriesId === series.seriesId)?.query ||
                    'QUERY NOT FOUND',
                stroke: insight.presentation.seriesPresentation.find(
                    presentation => presentation.seriesId === series.seriesId
                )?.color,
            }))

            if (isBackendInsight) {
                return {
                    type: InsightType.Backend,
                    presentationType: 'LineChartInsightViewPresentation',
                    id: insight.id,
                    visibility: '',
                    title: insight.presentation.title,
                    series,
                }
            }

            const repositories = uniq(
                insight.dataSeriesDefinitions.flatMap(series => series.repositoryScope.repositories)
            )

            const step = getDurationFromStep(insight.dataSeriesDefinitions[0].timeScope)

            return {
                type: InsightType.Extension,
                presentationType: 'LineChartInsightViewPresentation',
                id: insight.id,
                title: insight.presentation.title,
                visibility: '',
                step,
                repositories,
                series,
            }
        }
    }
}
