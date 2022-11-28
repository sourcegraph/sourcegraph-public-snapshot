import { Point } from '../types'

import { getDatumValue, isDatumWithValidNumber, SeriesWithData, SeriesDatum } from './data-series-processing'

const NULL_LINK = (): undefined => undefined

export function generatePointsField<Datum>(dataSeries: SeriesWithData<Datum>[]): { [seriesId: string]: Point[] } {
    const starter: { [key: string]: Point[] } = {}

    return dataSeries.reduce((previous, series) => {
        const { id, data, getLinkURL = NULL_LINK } = series

        previous[id] = (data as SeriesDatum<Datum>[]).filter(isDatumWithValidNumber).map((datum, index) => {
            const datumValue = getDatumValue(datum)

            return {
                id: getPointId(id, index),
                seriesId: id.toString(),
                yValue: datumValue,
                xValue: datum.x,
                color: series.color ?? 'green',
                linkUrl: getLinkURL(datum.datum, index),
            }
        })

        return previous
    }, starter)
}

export function getPointId(seriesId: string | number, pointIndex: number): string {
    return `${seriesId.toString()}${pointIndex}`
}
