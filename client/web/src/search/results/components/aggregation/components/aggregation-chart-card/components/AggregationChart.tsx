import { ReactElement, useMemo } from 'react'

import { ParentSize } from '@visx/responsive'
import { getTicks } from '@visx/scale'
import { AnyD3Scale } from '@visx/scale/lib/types/Scale'

import { SearchAggregationMode } from '@sourcegraph/shared/src/graphql-operations'
import { BarChart, BarChartProps } from '@sourcegraph/wildcard'

/**
 * AggregationChart sets these props internally, and we don't expose them
 * as public api of aggregation chart
 */
type PredefinedBarProps =
    | 'width'
    | 'height'
    | 'mode'
    | 'pixelsPerXTick'
    | 'pixelsPerYTick'
    | 'maxAngleXTick'
    | 'getScaleXTicks'
    | 'getTruncatedXTick'

type SharedBarProps<Datum> = Omit<BarChartProps<Datum>, PredefinedBarProps>

export interface AggregationChartProps<Datum> extends SharedBarProps<Datum> {
    mode?: SearchAggregationMode | null
    minAngleXTick: number
    maxXLabelLength: number
}

export function AggregationChart<Datum>(props: AggregationChartProps<Datum>): ReactElement {
    const { mode, minAngleXTick, maxXLabelLength, className, ...attributes } = props

    const getTruncatedXLabel = useMemo(() => getTruncationFormatter(mode, maxXLabelLength), [mode, maxXLabelLength])

    return (
        <ParentSize className={className}>
            {parent => (
                <BarChart
                    {...attributes}
                    sortByValue={true}
                    width={parent.width}
                    height={parent.height}
                    hideXTicks={true}
                    pixelsPerYTick={20}
                    pixelsPerXTick={20}
                    minAngleXTick={minAngleXTick}
                    maxAngleXTick={45}
                    getScaleXTicks={getXScaleTicks}
                    getTruncatedXTick={getTruncatedXLabel}
                />
            )}
        </ParentSize>
    )
}

interface GetScaleTicksOptions {
    scale: AnyD3Scale
    space: number
    pixelsPerTick?: number
}

/**
 * Custom tick generator for search result aggregation bar. Double down tick
 * labels count every until their count fits in a given available space.
 *
 * ```
 * Before
 * ─┬──┬──┬──┬──┬──┬──┬──┬──┬──┬───▶
 *  1  2  3  4  5  6  7  8  9  10
 *
 * After
 * ─┬─────┬─────┬─────┬─────┬──────▶
 *  1     3     5     7     9
 * ```
 */
const getXScaleTicks = <T,>(options: GetScaleTicksOptions): T[] => {
    const { scale, space, pixelsPerTick = 80 } = options

    // Calculate desirable number of ticks
    const numberTicks = Math.max(2, Math.floor(space / pixelsPerTick))

    let filteredTicks = getTicks(scale)

    while (filteredTicks.length > numberTicks) {
        filteredTicks = filteredTicks.filter((tick, index) => index % 2 === 0)
    }

    return filteredTicks
}

const getTruncatedTick = (maxLength: number) => (tick: string): string =>
    tick.length >= maxLength ? `${tick.slice(0, maxLength)}...` : tick

const getTruncatedTickFromTheEnd = (maxLength: number) => (tick: string): string =>
    tick.length >= maxLength ? `...${tick.slice(-maxLength)}` : tick

/**
 * Based on aggregation mode we should pick different truncation formatters for X labels.
 * Since Repo and FilePath aggregations usually have long labels with same sequence of symbols
 * at the beginning we truncate them from the end.
 *
 * ```
 * github.com/sourcegraph/about -> ...sourcegraph/about
 * github.com/sourcegraph/sourcegraph -> ...urcegraph/sourcegraph
 * ```
 */
const getTruncationFormatter = (
    aggregationMode: SearchAggregationMode | undefined | null,
    maxLength: number
): ((tick: string) => string) => {
    switch (aggregationMode) {
        // These types possible have long labels with the same pattern at the start of the string,
        // so we truncate their labels from the end
        case SearchAggregationMode.REPO:
        case SearchAggregationMode.PATH:
            return getTruncatedTickFromTheEnd(maxLength)

        default:
            return getTruncatedTick(maxLength)
    }
}
