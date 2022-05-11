import { ReactElement, useContext, useMemo } from 'react'

import { PatternLines } from '@visx/pattern'
import { DataContext } from '@visx/xychart'
import { LineChartContent } from 'sourcegraph'

import { Accessors, YAccessor } from '../types'

const patternId = 'xy-chart-pattern'

function getFirstNonNullablePoint<Datum>(data: Datum[], accessor: YAccessor<Datum>): Datum | undefined {
    return data.find(datum => accessor(datum) !== null)
}

export interface NonActiveBackgroundProps<Datum extends object, Key extends keyof Datum> {
    data: Datum[]
    series: LineChartContent<Datum, keyof Datum>['series']
    accessors: Accessors<Datum, Key>
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
 * */
export function NonActiveBackground<Datum extends object>(
    props: NonActiveBackgroundProps<Datum, keyof Datum>
): ReactElement | null {
    const { data, series, accessors } = props
    const { theme, margin, width, height, innerHeight, xScale } = useContext(DataContext)

    const backgroundWidth = useMemo(() => {
        // For non active background's width we need to find first non nullable element
        const firstNonNullablePoints: (Datum | undefined)[] = []

        for (const line of series) {
            const lineKey = line.dataKey
            const lineYAccessor = accessors.y[lineKey]

            firstNonNullablePoints.push(getFirstNonNullablePoint(data, lineYAccessor))
        }

        const lastNullablePointX = firstNonNullablePoints.reduce((xCoord, datum) => {
            // In case if the first non nullable element is the first element
            // of data that means we don't need to render non active background.
            if (!datum || datum === data[0]) {
                return null
            }

            // Get x point from datum by x accessor
            return accessors.x(datum)
        }, null as Date | number | null)

        // If we didn't find any non-nullable elements we don't need to render
        // non-active background.
        if (!lastNullablePointX) {
            return 0
        }

        // Convert x value of first non nullable point to reals svg coordinate
        const xValue = xScale?.(lastNullablePointX) ?? 0

        return +xValue
    }, [data, series, accessors, xScale])

    // Early return values not available in context or we don't need render
    // non active background.
    if (!backgroundWidth || !width || !height || !margin || !theme) {
        return null
    }

    return (
        <>
            <PatternLines
                id={patternId}
                width={16}
                height={16}
                orientation={['diagonal']}
                stroke={theme?.gridStyles?.stroke}
                strokeWidth={1}
            />

            <rect x={0} y={0} width={width} height={height} fill="transparent" />

            <rect
                x={margin.left}
                y={margin.top}
                width={backgroundWidth - margin.left}
                height={innerHeight}
                fill={`url(#${patternId})`}
                fillOpacity={0.3}
            />
        </>
    )
}
