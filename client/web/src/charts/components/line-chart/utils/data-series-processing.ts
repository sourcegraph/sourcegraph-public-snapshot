import { stack } from '@visx/shape'
import { Series } from 'd3-shape'

import { Series as ChartSeries } from '../../../types'

import { isValidNumber } from './data-guards'

export interface LineChartSeriesWithData<Datum> extends ChartSeries<Datum> {
    data: Datum[]
    originalData: Datum[]
    stackedSeries: Series<Datum, keyof Datum> | null
}

interface SeriesWithDataInput<Datum> {
    data: Datum[]
    series: ChartSeries<Datum>[]
    stacked: boolean
    xAxisKey: keyof Datum
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
export function getSeriesWithData<Datum>(input: SeriesWithDataInput<Datum>): LineChartSeriesWithData<Datum>[] {
    const { data, series, stacked, xAxisKey } = input

    if (!stacked) {
        return (
            series
                // Separate datum object by series lines
                .map<LineChartSeriesWithData<Datum>>(line => {
                    const filteredData = getFilteredSeriesData(data, datum => datum[line.dataKey])

                    return {
                        ...line,
                        // Filter select series data from the datum object and process this points array
                        data: filteredData,
                        originalData: filteredData,
                        stackedSeries: null,
                    }
                })
        )
    }

    const stackedDataMap: Record<string, Datum> = {}
    const stackedSeriesData = stack<Datum, keyof Datum>({
        keys: series.map(line => line.dataKey),
        order: 'ascending',
    })(data)

    for (const stackedSeries of stackedSeriesData) {
        for (const point of stackedSeries) {
            // D3-stack api feature. The stacked data has array shape where 0 indexed element
            // is a lower border for the stacked datum and 1 indexed element is a upper boundary
            // for the stacked line.
            const newStackedValue = point['1']
            const date = (point.data[xAxisKey] as unknown) as string

            if (!stackedDataMap[date]) {
                stackedDataMap[date] = { ...point.data }
            }

            stackedDataMap[date] = {
                ...stackedDataMap[date],
                [stackedSeries.key]: isValidNumber(point.data[stackedSeries.key]) ? newStackedValue : null,
            }
        }
    }

    const stackedDataList = [...Object.values(stackedDataMap)]

    return (
        series
            // Separate datum object by series lines
            .map<LineChartSeriesWithData<Datum>>((line, index) => ({
                ...line,
                // Filter select series data from the datum object and process this points array
                data: getFilteredSeriesData(stackedDataList, datum => datum[line.dataKey]),
                originalData: getFilteredSeriesData(data, datum => datum[line.dataKey]),
                stackedSeries: stackedSeriesData[index] ?? null,
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
