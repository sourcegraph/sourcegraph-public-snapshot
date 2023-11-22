import React, { memo, useMemo } from 'react'

import type { ScaleLinear, ScaleTime } from 'd3-scale'
import * as uuid from 'uuid'

import type { SeriesWithData } from '../../utils'

import { getStackedAreaPaths } from './get-stacked-area-paths'

interface StackedAreaProps<Datum> {
    dataSeries: SeriesWithData<Datum>[]
    yScale: ScaleLinear<number, number>
    xScale: ScaleTime<number, number>
}

/**
 * Draw stacked area paths (areas below the actual series line) for stacked
 * line chart's series.
 *
 * Example
 * ```
 *   ▲
 *   │              ●
 *   │        ● ▒▒▒▒▒▒▒▒▒
 *   │  ● ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒●▒▒▒▒▒▒●
 *   │  ▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒
 *   │  ▒▒▒▒▒▒▒▒▒▒▒▒◇▒▒▒▒▒▒ Series B area
 *   │  ▒▒▒▒▒▒◇░░░░░░░░░▒▒▒▒▒▒▒▒▒▒
 *   │  ◇░░░░░░░░░░░░░░░░░◇░░░░░░◇
 *   │  ░░░░░░░░░░░░░░░░░░░░░░░░░░
 *   │  ░░░░░░░░░░░░░░░░░░░ Series A area
 * ──┼─────────────────────────────────▶
 *   │
 * ```
 */
function StackedAreaInternal<Datum>(props: StackedAreaProps<Datum>): React.ReactElement {
    const { dataSeries, yScale, xScale } = props
    const id = useMemo(() => uuid.v4(), [])

    const seriesPaths = useMemo(() => getStackedAreaPaths({ dataSeries, yScale, xScale }), [dataSeries, yScale, xScale])

    return (
        <g>
            <defs>
                {seriesPaths.map(series => (
                    <path key={`stack-path-${id}-${series.id}`} id={`stack-path-${id}-${series.id}`} d={series.path} />
                ))}
                {seriesPaths
                    .filter((series, index) => index > 0)
                    .map((series, index) => (
                        <mask key={`mask-stack-${id}-${series.id}`} id={`mask-stack-${id}-${series.id}`}>
                            <rect width="100%" height="100%" fill="white" />

                            {
                                // This is safe because we filtered out the first elements
                                // of original seriesPaths array
                                seriesPaths.slice(0, index + 1).map(series => (
                                    <use key={series.id} href={`#stack-path-${id}-${series.id}`} fill="black" />
                                ))
                            }
                        </mask>
                    ))}
            </defs>

            {seriesPaths.map(series => (
                <use
                    key={`stack-path-${id}-${series.id}`}
                    href={`#stack-path-${id}-${series.id}`}
                    stroke="transparent"
                    opacity={0.5}
                    fill={series.color}
                    mask={`url(#mask-stack-${id}-${series.id})`}
                />
            ))}
        </g>
    )
}

const typedMemo: <T>(c: T) => T = memo
export const StackedArea = typedMemo(StackedAreaInternal)
