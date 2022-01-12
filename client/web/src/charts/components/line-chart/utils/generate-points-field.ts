import { ScaleLinear, ScaleTime } from 'd3-scale'

import { LineChartSeries, Point } from '../types'

import { isValidNumber } from './data-guards'

interface PointsFieldInput<Datum extends object> {
    data: Datum[]
    series: LineChartSeries<Datum>[]
    xScale: ScaleTime<number, number>
    yScale: ScaleLinear<number, number>
    xAxisKey: keyof Datum
}

export function generatePointsField<Datum extends object>(input: PointsFieldInput<Datum>): Point[] {
    const { data, series, xScale, yScale, xAxisKey } = input

    return data.flatMap((datum, index) =>
        series
            .filter(line => isValidNumber(datum[line.dataKey]))
            .map<Point>(line => ({
                id: `${line.dataKey as string}-${index}`,
                seriesKey: line.dataKey as string,
                value: +datum[line.dataKey],
                x: xScale(+datum[xAxisKey]),
                y: yScale(+datum[line.dataKey]),
                color: line.color ?? 'green',
                linkUrl: line.linkURLs?.[index],
                index,
            }))
    )
}
