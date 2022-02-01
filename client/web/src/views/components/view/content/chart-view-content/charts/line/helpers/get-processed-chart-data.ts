import { LineChartSeries } from 'sourcegraph'

import { Accessors, Point, LineChartSeriesWithData } from '../types'

interface GetProcessedChartDataProps<Datum extends object> {
    data: Datum[]

    /**
     * Accessors map to get (x, y) value from datum objects
     */
    accessors: Accessors<Datum, keyof Datum>

    series: LineChartSeries<Datum>[]
}

interface ProcessedChartData<Datum extends object> {
    /**
     * List of processed data series data for each data series line.
     */
    seriesWithData: LineChartSeriesWithData<Datum>[]

    sortedData: Datum[]
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
export function getProcessedChartData<Datum extends object>(
    props: GetProcessedChartDataProps<Datum>
): ProcessedChartData<Datum> {
    const { data, series, accessors } = props

    // In case if we've got unsorted by x (time) axis dataset we have to sort that by ourselves
    // otherwise we will get an error in calculation of position for the tooltip
    // Details: bisector from d3-array package expects sorted data otherwise he can't calculate
    // right index for nearest point on the XYChart.
    // See https://github.com/airbnb/visx/blob/master/packages/visx-xychart/src/utils/findNearestDatumSingleDimension.ts#L30
    const sortedData = data.sort((firstDatum, secondDatum) => +accessors.x(firstDatum) - +accessors.x(secondDatum))

    const seriesWithData = series
        // Separate datum object by series lines
        .map<LineChartSeriesWithData<Datum>>(line => ({
            ...line,
            // Filter select series data from the datum object and process this points array
            data: getFilteredSeriesData(
                sortedData.map(datum => ({
                    x: accessors.x(datum),
                    y: accessors.y[line.dataKey](datum),
                }))
            ),
        }))

    return { sortedData, seriesWithData }
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
 *
 * @param data - Series data list
 */
function getFilteredSeriesData<Datum>(data: Point[]): Point[] {
    const firstNonNullablePointIndex = Math.max(
        data.findIndex(datum => datum.y !== null),
        0
    )

    // Preserve null values at the beginning of the series data list
    // but remove null holes between the points further.
    const nullBeginningValues = data.slice(0, firstNonNullablePointIndex)
    const pointsWithoutHoles = data
        // Get values after null area
        .slice(firstNonNullablePointIndex)
        .filter(point => point.y !== null)

    return [...nullBeginningValues, ...pointsWithoutHoles]
}
