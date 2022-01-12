import { curveLinear } from '@visx/curve'
import { Group } from '@visx/group'
import { scaleTime, scaleLinear } from '@visx/scale'
import { LinePath } from '@visx/shape'
import { voronoi } from '@visx/voronoi'
import classNames from 'classnames'
import { noop } from 'lodash'
import React, { ReactElement, useMemo, useRef, useState } from 'react'

import { AxisBottom, AxisLeft } from './components/axis/Axis'
import { NonActiveBackground } from './components/NonActiveBackground'
import { PointGlyph } from './components/PointGlyph'
import { Tooltip, TooltipContent } from './components/tooltip/Tooltip'
import { useChartEventHandlers } from './hooks/event-listeners'
import styles from './LineChart.module.scss'
import { LineChartSeries, Point } from './types'
import { isValidNumber } from './utils/data-guards'
import { getSeriesWithData } from './utils/data-series-processing'
import { generatePointsField } from './utils/generate-points-field'
import { getChartContentSizes } from './utils/get-chart-content-sizes'
import { getMinMaxBoundaries } from './utils/get-min-max-boundary'

export interface LineChartContentProps<D extends object> {
    width: number
    height: number

    /** An array of data objects, with one element for each step on the X axis. */
    data: D[]

    /** The series (lines) of the chart. */
    series: LineChartSeries<D>[]

    /**
     * The key in each data object for the X value this line should be
     * calculated from.
     */
    xAxisKey: keyof D

    /**
     * Callback runs whenever a point-zone (zone around point) and point itself
     * on the chart is clicked.
     */
    onDatumClick?: (event: React.MouseEvent) => void
}

/**
 * Visual component that renders svg line chart with pre-defined sizes, tooltip,
 * voronoi area distribution.
 */
export function LineChart<D extends object>(props: LineChartContentProps<D>): ReactElement | null {
    const { width: outerWidth, height: outerHeight, data, series, xAxisKey, onDatumClick = noop } = props

    const [activePoint, setActivePoint] = useState<Point & { element?: Element }>()
    const yAxisReference = useRef<SVGGElement>(null)
    const xAxisReference = useRef<SVGGElement>(null)

    const { width, height, margin } = useMemo(
        () =>
            getChartContentSizes({
                width: outerWidth,
                height: outerHeight,
                margin: {
                    top: 10,
                    right: 10,
                    left: yAxisReference.current?.getBoundingClientRect().width ?? 30,
                    bottom: xAxisReference.current?.getBoundingClientRect().height ?? 30,
                },
            }),
        [outerWidth, outerHeight, yAxisReference]
    )

    const { minX, maxX, minY, maxY } = useMemo(() => getMinMaxBoundaries({ data, series, xAxisKey }), [
        data,
        series,
        xAxisKey,
    ])

    const xScale = useMemo(
        () =>
            scaleTime({
                domain: [minX, maxX],
                range: [margin.left, width],
                nice: true,
            }),
        [minX, maxX, margin.left, width]
    )

    const yScale = useMemo(
        () =>
            scaleLinear({
                domain: [minY, maxY],
                range: [height, margin.top],
                nice: true,
            }),
        [minY, maxY, margin.top, height]
    )

    const points = useMemo(() => generatePointsField({ data, series, xAxisKey, yScale, xScale }), [
        data,
        series,
        xAxisKey,
        yScale,
        xScale,
    ])

    const dataSeries = useMemo(() => getSeriesWithData({ data, series }), [data, series])

    const voronoiLayout = useMemo(
        () =>
            voronoi<Point>({
                x: point => point.x,
                y: point => point.y,
                width,
                height,
            })(points),
        [width, height, points]
    )

    const handlers = useChartEventHandlers({
        onPointerMove: point => {
            const closestPoint = voronoiLayout.find(point.x, point.y)

            if (closestPoint && closestPoint.data.id !== activePoint?.id) {
                setActivePoint(closestPoint.data)
            }
        },
        onPointerLeave: () => setActivePoint(undefined),
        onClick: event => {
            if (activePoint?.linkUrl) {
                onDatumClick(event)
                window.open(activePoint.linkUrl)
            }
        },
    })

    return (
        <svg
            width={outerWidth}
            height={outerHeight}
            className={classNames(styles.root, { [styles.rootWithHoveredLinkPoint]: activePoint?.linkUrl })}
            {...handlers}
        >
            <AxisLeft
                ref={yAxisReference}
                scale={yScale}
                width={width}
                height={height}
                top={margin.top}
                left={margin.left}
            />

            <AxisBottom ref={xAxisReference} scale={xScale} top={margin.top + height} width={width} />

            <NonActiveBackground
                data={data}
                series={series}
                xAxisKey={xAxisKey}
                width={width}
                height={height}
                top={margin.top}
                left={margin.left}
                xScale={xScale}
            />

            <Group top={margin.top}>
                {dataSeries.map(line => (
                    <LinePath
                        key={line.dataKey as string}
                        data={line.data}
                        curve={curveLinear}
                        defined={datum => isValidNumber(datum[line.dataKey])}
                        x={datum => xScale(+datum[xAxisKey])}
                        y={datum => yScale(+datum[line.dataKey])}
                        stroke={line.color}
                        strokeWidth={2}
                        strokeLinecap="round"
                    />
                ))}

                {points.map(point => (
                    <PointGlyph
                        key={point.id}
                        left={point.x}
                        top={point.y}
                        active={activePoint?.id === point.id}
                        color={point.color}
                        linkURL={point.linkUrl}
                        onClick={onDatumClick}
                        onFocus={event => setActivePoint({ ...point, element: event.target })}
                        onBlur={() => setActivePoint(undefined)}
                    />
                ))}
            </Group>

            {activePoint && (
                <Tooltip reference={activePoint.element}>
                    <TooltipContent data={data} series={series} xAxisKey={xAxisKey} activePoint={activePoint} />
                </Tooltip>
            )}
        </svg>
    )
}
