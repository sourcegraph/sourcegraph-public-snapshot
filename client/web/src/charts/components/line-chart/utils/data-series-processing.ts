import { LineChartSeries } from '../types'

import { isValidNumber } from './data-guards'

export interface LineChartSeriesWithData<Datum> extends LineChartSeries<Datum> {
    data: Datum[]
}

interface SeriesWithDataInput<Datum extends object> {
    data: Datum[]
    series: LineChartSeries<Datum>[]
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
export function getSeriesWithData<Datum extends object>(
    input: SeriesWithDataInput<Datum>
): LineChartSeriesWithData<Datum>[] {
    const { data, series } = input

    return (
        series
            // Separate datum object by series lines
            .map<LineChartSeriesWithData<Datum>>(line => ({
                ...line,
                // Filter select series data from the datum object and process this points array
                data: getFilteredSeriesData(data, datum => datum[line.dataKey]),
            }))
    )
}

/**
 * Filters series data list, preserves null value at the beginning of the series data list
 * and removes null value between the points.
 *
 * ```
 * Null value ▽   Real point ■                  Null value ▽   Real point ■
 * ┌────────────────────────────────────┐       ┌────────────────────────────────────┐
 * │░░░░░░░░░░░░░░░                     │       │░░░░░░░░░░░░░░░                     │
 * │░░░░░░░░░░░░░░░                     │       │░░░░░░░░░░░░░░░                     │
 * │░░░░░░░░░░░░░░░                ■    │       │░░░░░░░░░░░░░░░                ■    │
 * │░░░░░░░░░░░░▽░░    ■                │       │░░░░░░░░░░░░▽░░    ■                │
 * │░░░░░░░░░░░░░░░          ▽          │──────▶│░░░░░░░░░░░░░░░                     │
 * │░░░░░░▽░░░░░░░░ ■                   │       │░░░░░░▽░░░░░░░░ ■                   │
 * │░░░░░░░░░░░░░░░       ■             │       │░░░░░░░░░░░░░░░       ■             │
 * │░░░▽░░░░░░░░░░░                     │       │░░░▽░░░░░░░░░░░                     │
 * │░░░░░░░░░░░░░░░             ▽       │       │░░░░░░░░░░░░░░░                     │
 * └────────────────────────────────────┘       └────────────────────────────────────┘
 *```
 */
function getFilteredSeriesData<Datum>(data: Datum[], yAccessor: (d: Datum) => unknown): Datum[] {
    const firstNonNullablePointIndex = Math.max(
        data.findIndex(datum => isValidNumber(yAccessor(datum))),
        0
    )

    // Preserve null values at the beginning of the series data list
    // but remove null holes between the points further.
    const nullBeginningValues = data.slice(0, firstNonNullablePointIndex)
    const pointsWithoutHoles = data
        // Get values after null area
        .slice(firstNonNullablePointIndex)
        .filter(datum => isValidNumber(yAccessor(datum)))

    return [...nullBeginningValues, ...pointsWithoutHoles]
}
