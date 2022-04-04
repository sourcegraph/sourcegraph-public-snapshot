import { Series } from '../../../../types'
import { isValidNumber } from '../data-guards'

import { getStackedSeriesData } from './get-stacked-series-data'
import { getFilteredSeriesData } from './helpers'
import { IndependentSeries, SeriesType, SeriesWithData, StandardSeriesDatum } from './types'

interface Input<Datum> {
    data: Datum[]
    series: Series<Datum>[]
    stacked: boolean
    getXValue: (datum: Datum) => Date
}

/**
 * Processes data list (sort it) and extracts data from the datum object and merge it with
 * series (line) data.
 *
 * Example:
 * ```
 * data: [
 *   { a: 2, b: 3, x: 2},
 *   { a: 4, b: 5, x: 3},
 *   { a: null, b: 6, x: 4}
 * ] → series: [
 *     { name: a, data: [{ y: 2, x: 2 }, { y: 4, x: 3 }],
 *     { name: b, data: [{ y: 3, x: 2 }, { y: 5, x: 3 }, { y: 6, x: 4 }]
 *   ]
 * ```
 */
export function getSeriesData<Datum>(input: Input<Datum>): SeriesWithData<Datum>[] {
    const { data, series, stacked, getXValue } = input

    if (!stacked) {
        return (
            series
                // Separate datum object by series lines
                .map<IndependentSeries<Datum>>(line => {
                    const lineData = data.map<StandardSeriesDatum<Datum>>(datum => ({
                        datum,
                        x: getXValue(datum),
                        y: isValidNumber(datum[line.dataKey]) ? +datum[line.dataKey] : null,
                    }))

                    return {
                        ...line,
                        type: SeriesType.Independent,
                        // Filter select series data from the datum object and process this points array
                        data: getFilteredSeriesData(lineData) as StandardSeriesDatum<Datum>[],
                    }
                })
        )
    }

    return getStackedSeriesData(input)
}
