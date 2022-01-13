import { AxisScale } from '@visx/axis/lib/types'
import { PatternLines } from '@visx/pattern'
import React, { ReactElement, useMemo } from 'react'

import { LineChartSeries } from '../types'
import { isValidNumber } from '../utils/data-guards'

const PATTERN_ID = 'xy-chart-pattern'

export interface NonActiveBackgroundProps<Datum extends object> {
    data: Datum[]
    series: LineChartSeries<Datum>[]
    xAxisKey: keyof Datum
    xScale: AxisScale
    width: number
    height: number
    left: number
    top: number
}

/**
 * Displays custom pattern background for area where we don't have any data points.
 * Example:
 * ┌──────────────────────────────────┐
 * │``````````````````                │ 10
 * │``````````````````                │
 * │``````````````````              ▼ │ 9
 * │``````````````````                │
 * │``````````````````      ▼         │ 8
 * │``````````````````                │
 * │``````````````````          ▼     │ 7
 * │`````````````````` ▼              │
 * │``````````````````                │ 6
 * │``````````````````                │
 * │``````````````````                │ 5
 * └──────────────────────────────────┘
 * Where ` is a non-active background
 */
export function NonActiveBackground<Datum extends object>(props: NonActiveBackgroundProps<Datum>): ReactElement | null {
    const { data, series, xAxisKey, xScale, top, left, width, height } = props

    const backgroundWidth = useMemo(() => {
        // For non active background's width we need to find first non nullable element
        const firstNonNullablePoints: (Datum | undefined)[] = series.map(line =>
            data.find(datum => isValidNumber(datum[line.dataKey]))
        )

        const lastNullablePointX = firstNonNullablePoints.reduce((xCoord, datum) => {
            // In case if the first non nullable element is the first element
            // of data that means we don't need to render non active background.
            if (!datum || datum === data[0]) {
                return null
            }

            // Get x point from datum by x accessor
            return +datum[xAxisKey]
        }, null as Date | number | null)

        // If we didn't find any non-nullable elements we don't need to render
        // non-active background.
        if (!lastNullablePointX) {
            return 0
        }

        // Convert x value of first non nullable point to reals svg coordinate
        const xValue = xScale?.(lastNullablePointX) ?? 0

        return +xValue
    }, [data, series, xAxisKey, xScale])

    // Early return values not available in context or we don't need render
    // non active background.
    if (!backgroundWidth || !width || !height) {
        return null
    }

    return (
        <>
            <PatternLines
                id={PATTERN_ID}
                width={16}
                height={16}
                orientation={['diagonal']}
                stroke="var(--border-color)"
                strokeWidth={2}
            />

            <rect x={left} y={top} width={backgroundWidth - left} height={height} fill={`url(#${PATTERN_ID})`} />
        </>
    )
}
