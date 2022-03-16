import { ScaleLinear, ScaleTime } from 'd3-scale'

import { isDefined } from '@sourcegraph/common'

import { Point } from '../types'

import { isValidNumber } from './data-guards'
import { LineChartSeriesWithData } from './data-series-processing'

interface PointsFieldInput<Datum> {
    dataSeries: LineChartSeriesWithData<Datum>[]
    xScale: ScaleTime<number, number>
    yScale: ScaleLinear<number, number>
    xAxisKey: keyof Datum
}

export function generatePointsField<Datum>(input: PointsFieldInput<Datum>): Point<Datum>[] {
    const { dataSeries, xScale, yScale, xAxisKey } = input

    return dataSeries.flatMap(series =>
        series.data
            .map((datum, index) =>
                isValidNumber(datum[series.dataKey])
                    ? {
                          id: `${series.dataKey as string}-${index}`,
                          seriesKey: series.dataKey as string,
                          value: +datum[series.dataKey],
                          x: xScale(+datum[xAxisKey]),
                          y: yScale(+datum[series.dataKey]),
                          color: series.color ?? 'green',
                          linkUrl: series.linkURLs?.[index],
                          originalDatum: series.originalData[index],
                          datum,
                      }
                    : null
            )
            .filter(isDefined)
    )
}
