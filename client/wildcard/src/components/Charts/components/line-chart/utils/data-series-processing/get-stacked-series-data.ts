import { scaleTime } from '@visx/scale'

import type { Series as ChartSeries } from '../../../../types'
import { isValidNumber } from '../data-guards'

import { encodePointId } from './helpers'
import { SeriesType, type StackedSeries, type StackedSeriesDatum } from './types'

/**
 * Iterate over data and series and tries to "stack" series values on each other. So basically
 * this implements sum operation over series data.
 *
 * ```
 *   ▲                    ◇ Series A area    ▲
 *   │                    ● Series B area    │
 *   │                                       │  ●                        ●
 *   │  ●                                    │  ▲     ●                  ▲
 *   │        ●                  ●           │  │     ▲     ●     ●      │
 *   │              ◇     ●      ◇           │  │     │     ◇     ▲      ◇
 *   │        ◇                              │  │     ◇           │
 *   │  ◇           ●     ◇                  │  ◇                 ◇
 * ──┼──────────────────────────────▶      ──┼──────────────────────────────▶
 * ```
 */
export function getStackedSeriesData<Datum>(series: ChartSeries<Datum>[]): StackedSeries<Datum>[] {
    const stack: StackedSeries<Datum>[] = []

    // eslint-disable-next-line ban/ban
    series.forEach(line => {
        const { data, getXValue, getYValue, id } = line
        const stackedData: StackedSeriesDatum<Datum>[] = []

        for (const [index, datum] of data.entries()) {
            const value = getYValue(datum)

            if (isValidNumber(value)) {
                const previousValue = findPreviousValueOnStack(stack, getXValue(datum), index)

                stackedData.push({
                    datum,
                    id: encodePointId(id, index),
                    y0: previousValue,
                    y1: previousValue + value,
                    x: getXValue(datum),
                })
            } else {
                stackedData.push({
                    datum,
                    id: `invalid-datum-${index}`,
                    y0: null,
                    y1: null,
                    x: getXValue(datum),
                })
            }
        }

        stack.push({
            ...line,
            type: SeriesType.Stacked,
            data: stackedData,
        })
    })

    return stack
}

/**
 * Sometimes series are not aligned to each other by x axis and in this case
 * for example for series B (on the schema picture below) in x2 time point
 * doesn't have an explicit base value from series below (series A). So
 * in this case we have to get points x1 and x3 of series A, and interpolate
 * they values to figure out value for x2 time.
 * ```
 *               ┌── Point without base value below
 *   ▲           │
 *   │        ● ◀┘
 *   │        │         ●
 *   │  ●               ▲
 *   │  ▲     │         │
 *   │  │               ◇    ◇ Series A
 *   │  ◇─ ─ ─│─ ─ ◇    │    ● Series B
 *   │  │          │
 * ──┼────────┴─────────┴───────────▶
 *   │  x1    x2   x3   x4
 * ```
 */
function findPreviousValueOnStack<Datum>(
    stack: StackedSeries<Datum>[],
    wantedDate: Date,
    currentIndex: number
): number {
    for (let index = stack.length - 1; index >= 0; index--) {
        const stackedSeries = stack[index]

        // Try to find stack value by datum index - if lines has the same points
        // in same x points it should match in just O(1)
        const stackedExactDatum = stackedSeries.data[currentIndex]
        const hasExactMatch = +stackedExactDatum?.x === +wantedDate

        if (hasExactMatch && stackedExactDatum.y1 !== null) {
            return stackedExactDatum.y1
        }

        const stackedDatum = stackedSeries.data.find(
            stackedDatum => stackedDatum.x === wantedDate && stackedDatum.y1 !== null
        )

        if (stackedDatum) {
            return stackedDatum.y1 as number
        }

        const interpolateValue = getInterpolatedValue(stackedSeries, wantedDate)

        if (interpolateValue !== null) {
            return interpolateValue
        }
    }

    return 0
}

function getInterpolatedValue<Datum>(stackedSeries: StackedSeries<Datum>, wantedDate: Date): number | null {
    const data = stackedSeries.data.filter(datum => datum.y1 !== null)

    for (let index = 0; index <= data.length - 2; index++) {
        const currentDatum = data[index]
        const nextDatum = data[index + 1]

        if (currentDatum.y1 === null || nextDatum.y1 === null) {
            continue
        }

        if (currentDatum.x <= wantedDate && wantedDate <= nextDatum.x) {
            const scale = scaleTime({
                domain: [new Date(+currentDatum.x), new Date(+nextDatum.x)],
                range: [0, 1],
            })

            const progress = scale(wantedDate)

            // Interpolate between two values a * (1 - t) + b * t;]
            return currentDatum.y1 * (1 - progress) + nextDatum.y1 * progress
        }
    }

    return null
}
