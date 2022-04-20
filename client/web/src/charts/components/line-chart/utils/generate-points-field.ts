import { ScaleLinear, ScaleTime } from 'd3-scale'

import { Point } from '../types'

import { getDatumValue, isDatumWithValidNumber, SeriesWithData } from './data-series-processing'
import { SeriesDatum } from './data-series-processing/types'

const NULL_LINK = (): undefined => undefined

interface PointsFieldInput<Datum> {
    dataSeries: SeriesWithData<Datum>[]
    xScale: ScaleTime<number, number>
    yScale: ScaleLinear<number, number>
    getXValue: (datum: Datum) => Date
}

export function generatePointsField<Datum>(input: PointsFieldInput<Datum>): Point<Datum>[] {
    const { dataSeries, xScale, yScale } = input

    return dataSeries.flatMap(series => {
        const { data, dataKey, getLinkURL = NULL_LINK } = series

        return (data as SeriesDatum<Datum>[]).filter(isDatumWithValidNumber).map((datum, index) => {
            const datumValue = getDatumValue(datum)

            return {
                id: `${dataKey as string}-${index}`,
                seriesKey: dataKey as string,
                value: datumValue,
                y: yScale(datumValue),
                x: xScale(datum.x),
                time: datum.x,
                color: series.color ?? 'green',
                linkUrl: getLinkURL(datum.datum, index),
                datum: datum.datum,
            }
        })
    })
}
