import { ScaleLinear, ScaleTime } from 'd3-scale'

import { isDefined } from '@sourcegraph/common'

import { Point } from '../types'

import { isValidNumber } from './data-guards'
import { LineChartSeriesWithData } from './data-series-processing'

const NULL_LINK = (): undefined => undefined

interface PointsFieldInput<Datum> {
    dataSeries: LineChartSeriesWithData<Datum>[]
    xScale: ScaleTime<number, number>
    yScale: ScaleLinear<number, number>
    xAxisKey: keyof Datum
}

export function generatePointsField<Datum>(input: PointsFieldInput<Datum>): Point<Datum>[] {
    const { dataSeries, xScale, yScale, xAxisKey } = input

    return dataSeries
        .flatMap(series => {
            const { dataKey, originalData, getLinkURL = NULL_LINK } = series

            return series.data.map((datum, index) =>
                isValidNumber(datum[dataKey])
                    ? {
                          id: `${dataKey as string}-${index}`,
                          seriesKey: dataKey as string,
                          value: +datum[dataKey],
                          x: xScale(+datum[xAxisKey]),
                          y: yScale(+datum[series.dataKey]),
                          color: series.color ?? 'green',
                          linkUrl: getLinkURL(datum),
                          originalDatum: originalData[index],
                          datum,
                      }
                    : null
            )
        })
        .filter(isDefined)
}
