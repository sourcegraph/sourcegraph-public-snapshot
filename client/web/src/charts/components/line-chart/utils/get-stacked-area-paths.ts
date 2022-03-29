import { area } from '@visx/shape'
import { ScaleLinear, ScaleTime } from 'd3-scale'
import { SeriesPoint } from 'd3-shape'

import { Series } from '../../../types'

import { isValidNumber } from './data-guards'
import { LineChartSeriesWithData } from './data-series-processing'

interface Props<Datum> {
    data: Datum[]
    dataSeries: LineChartSeriesWithData<Datum>[]
    xKey: keyof Datum
    yScale: ScaleLinear<number, number>
    xScale: ScaleTime<number, number>
}

interface SeriesPath<Datum> extends Series<Datum> {
    path: string
}

export function getStackedAreaPaths<Datum>(props: Props<Datum>): SeriesPath<Datum>[] {
    const { dataSeries, xKey, yScale, xScale } = props

    const paths: SeriesPath<Datum>[] = []

    for (const series of dataSeries) {
        const { stackedSeries } = series

        if (!stackedSeries) {
            continue
        }

        const path = area<SeriesPoint<Datum>>({
            x: seriesDatum => xScale(+seriesDatum.data[xKey]) ?? 0,
            y0: seriesDatum => yScale(seriesDatum[0]) ?? 0,
            y1: seriesDatum => yScale(seriesDatum[1]) ?? 0,
        })

        paths.push({
            ...series,
            path: path(stackedSeries.filter(series => isValidNumber(series.data[stackedSeries.key]))) ?? '',
        })
    }

    return paths
}
