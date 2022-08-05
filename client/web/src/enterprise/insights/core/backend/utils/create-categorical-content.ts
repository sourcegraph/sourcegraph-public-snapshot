import { groupBy } from 'lodash'

import { InsightDataSeries } from '../../../../../graphql-operations'
import { CategoricalChartContent } from '../code-insights-backend-types'

import { DATA_SERIES_COLORS_LIST } from './create-line-chart-content'

interface CategoricalDatum {
    name: string
    fill: string
    value: number
}

export function createCategoricalChart(seriesData: InsightDataSeries[]): CategoricalChartContent<CategoricalDatum> {
    const seriesGroups = groupBy(
        seriesData.filter(series => series.label),
        series => series.label
    )

    // Group series result by seres name and sum up series value with the same name
    const groups = Object.keys(seriesGroups).map((key, index) =>
        seriesGroups[key].reduce(
            (memo, series) => {
                memo.value += series.points.reduce((sum, datum) => sum + datum.value, 0)

                return memo
            },
            {
                name: seriesGroups[key][0].label,
                fill: DATA_SERIES_COLORS_LIST[index % DATA_SERIES_COLORS_LIST.length],
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
