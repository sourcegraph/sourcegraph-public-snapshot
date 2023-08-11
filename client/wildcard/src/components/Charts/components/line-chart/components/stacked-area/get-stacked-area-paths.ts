import { area } from '@visx/shape'
import type { ScaleLinear, ScaleTime } from 'd3-scale'

import { isDatumWithValidNumber, SeriesType, type SeriesWithData } from '../../utils'
import type { StackedSeriesDatum } from '../../utils/data-series-processing/types'

interface Props<Datum> {
    dataSeries: SeriesWithData<Datum>[]
    yScale: ScaleLinear<number, number>
    xScale: ScaleTime<number, number>
}

interface SeriesPath {
    id: string | number
    path: string
    color: string
}

export function getStackedAreaPaths<Datum>(props: Props<Datum>): SeriesPath[] {
    const { dataSeries, yScale, xScale } = props

    const paths: SeriesPath[] = []
    const bottomLine = yScale(0)

    for (const series of dataSeries) {
        if (series.type === SeriesType.Independent) {
            continue
        }

        const path = area<StackedSeriesDatum<Datum>>({
            x: point => xScale(new Date(+point.x)),
            y0: () => bottomLine,
            y1: point => yScale(point.y1 ?? 0),
        })

        const stackedData = series.data.filter(isDatumWithValidNumber)

        paths.push({
            id: series.id,
            color: series.color ?? 'gray',
            path: path(stackedData) ?? '',
        })
    }

    return paths
}
