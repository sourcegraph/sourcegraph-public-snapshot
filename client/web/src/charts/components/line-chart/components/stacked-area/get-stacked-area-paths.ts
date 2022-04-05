import { area } from '@visx/shape'
import { ScaleLinear, ScaleTime } from 'd3-scale'

import { Series } from '../../../../types'
import { isDatumWithValidNumber, SeriesType, SeriesWithData } from '../../utils'
import { StackedSeriesDatum } from '../../utils/data-series-processing/types'

interface Props<Datum> {
    dataSeries: SeriesWithData<Datum>[]
    yScale: ScaleLinear<number, number>
    xScale: ScaleTime<number, number>
}

interface SeriesPath<Datum> extends Series<Datum> {
    path: string
}

export function getStackedAreaPaths<Datum>(props: Props<Datum>): SeriesPath<Datum>[] {
    const { dataSeries, yScale, xScale } = props

    const paths: SeriesPath<Datum>[] = []
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
            ...series,
            path: path(stackedData) ?? '',
        })
    }

    return paths
}
