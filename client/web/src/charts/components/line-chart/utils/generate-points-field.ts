import { ScaleLinear, ScaleTime } from 'd3-scale'

import { Point } from '../types'

import { getDatumValue, isDatumWithValidNumber, SeriesWithData, SeriesDatum } from './data-series-processing'

const NULL_LINK = (): undefined => undefined

interface PointsFieldInput<Datum> {
    dataSeries: SeriesWithData<Datum>[]
    xScale: ScaleTime<number, number>
    yScale: ScaleLinear<number, number>
}

export function generatePointsField<Datum>(input: PointsFieldInput<Datum>): { [seriesId: string]: Point<Datum>[] } {
    const { dataSeries, xScale, yScale } = input
    const starter: { [key: string]: Point<Datum>[] } = {}

    return dataSeries
        .flatMap(series => {
            const { id, data, getLinkURL = NULL_LINK } = series

            return (data as SeriesDatum<Datum>[]).filter(isDatumWithValidNumber).map((datum, index) => {
                const datumValue = getDatumValue(datum)

                return {
                    id: `${id}-${index}`,
                    seriesId: id.toString(),
                    value: datumValue,
                    time: datum.x,
                    y: yScale(datumValue),
                    x: xScale(datum.x),
                    color: series.color ?? 'green',
                    linkUrl: getLinkURL(datum.datum, index),
                }
            })
        })
        .reduce((previous, current) => {
            previous[current.seriesId] = previous[current.seriesId] || []
            previous[current.seriesId].push(current)
            return previous
        }, starter)
}
