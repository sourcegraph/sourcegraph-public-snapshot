import type { Point } from '../types'

import { getDatumValue, isDatumWithValidNumber, type SeriesWithData, type SeriesDatum } from './data-series-processing'

const NULL_LINK = (): undefined => undefined

interface FlatSeriesMap {
    [seriesId: string]: Point[]
}

export function generatePointsField<Datum>(dataSeries: SeriesWithData<Datum>[]): FlatSeriesMap {
    const starter: { [key: string]: Point[] } = {}

    return dataSeries.reduce((previous, series) => {
        const { id, data, getLinkURL = NULL_LINK } = series

        previous[id] = (data as SeriesDatum<Datum>[]).filter(isDatumWithValidNumber).map((datum, index) => {
            const datumValue = getDatumValue(datum)

            return {
                id: datum.id,
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
