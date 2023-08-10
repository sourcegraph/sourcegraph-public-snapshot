import { groupBy } from 'lodash'

import type { InsightDataSeries } from '../../../../../graphql-operations'
import type { ComputeInsight } from '../../types'
import type { CategoricalChartContent } from '../code-insights-backend-types'

interface CategoricalDatum {
    name: string
    fill: string
    value: number
}

export function createComputeCategoricalChart(
    insight: ComputeInsight,
    seriesData: InsightDataSeries[]
): CategoricalChartContent<CategoricalDatum> {
    const seriesGroups = groupBy(
        seriesData.filter(series => series.label && series.points.length),
        series => series.label
    )

    // Group series result by seres name and sum up series value with the same name
    const groups = Object.keys(seriesGroups).map(key =>
        seriesGroups[key].reduce(
            (memo, series) => {
                memo.value += series.points.reduce((sum, datum) => sum + datum.value, 0)

                return memo
            },
            {
                name: seriesGroups[key][0].label,
                fill: insight.series[0]?.stroke ?? 'gray',
                value: 0,
            }
        )
    )

    return {
        data: groups,
        getDatumValue: datum => datum.value,
        getDatumColor: datum => datum.fill,
        getDatumName: datum => datum.name,
    }
}
