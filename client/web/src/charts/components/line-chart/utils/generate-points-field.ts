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

        return (data as SeriesDatum<Datum>[]).filter(isDatumWithValidNumber).map((data, index) => {
            const datumValue = getDatumValue(data)

            return {
                id: `${dataKey as string}-${index}`,
                seriesKey: dataKey as string,
                value: datumValue,
                y: yScale(datumValue),
                x: xScale(data.x),
                time: data.x,
                color: series.color ?? 'green',
                linkUrl: getLinkURL(data.datum),
                datum: data.datum,
            }
        })
    })
}
