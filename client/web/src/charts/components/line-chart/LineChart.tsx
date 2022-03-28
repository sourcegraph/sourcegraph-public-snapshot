import React, { ReactElement, useMemo, useRef, useState } from 'react'

import { curveLinear } from '@visx/curve'
import { Group } from '@visx/group'
import { scaleTime, scaleLinear } from '@visx/scale'
import { LinePath } from '@visx/shape'
import { voronoi } from '@visx/voronoi'
import classNames from 'classnames'
import { noop } from 'lodash'

import { SeriesLikeChart } from '../../types'

import { AxisBottom, AxisLeft, Tooltip, TooltipContent, NonActiveBackground, PointGlyph } from './components'
import { useChartEventHandlers } from './hooks/event-listeners'
import { Point } from './types'
import {
    isValidNumber,
    getSeriesWithData,
    generatePointsField,
    getChartContentSizes,
    getMinMaxBoundaries,
    getStackedAreaPaths,
} from './utils'

import styles from './LineChart.module.scss'

export interface LineChartContentProps<Datum> extends SeriesLikeChart<Datum> {
    width: number
    height: number
}

/**
 * Visual component that renders svg line chart with pre-defined sizes, tooltip,
 * voronoi area distribution.
 */
export function LineChart<D>(props: LineChartContentProps<D>): ReactElement | null {
    const {
        width: outerWidth,
        height: outerHeight,
        data,
        series,
        xAxisKey,
        stacked = false,
        onDatumClick = noop,
    } = props

    const [activePoint, setActivePoint] = useState<Point<D> & { element?: Element }>()
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

    const dataSeries = useMemo(() => getSeriesWithData({ data, series, stacked, xAxisKey }), [
        data,
        series,
        stacked,
        xAxisKey,
    ])

    const { minX, maxX, minY, maxY } = useMemo(() => getMinMaxBoundaries({ dataSeries, xAxisKey }), [
        dataSeries,
        xAxisKey,
    ])

    const xScale = useMemo(
        () =>
            scaleTime({
                domain: [minX, maxX],
                range: [margin.left, width],
                nice: true,
                clamp: true,
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

    const points = useMemo(() => generatePointsField({ dataSeries, xAxisKey, yScale, xScale }), [
        dataSeries,
        xAxisKey,
        yScale,
        xScale,
    ])

    const voronoiLayout = useMemo(
        () =>
            voronoi<Point<D>>({
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
                {stacked && (
                    <Group>
                        {getStackedAreaPaths({ data, dataSeries, xScale, yScale, xKey: xAxisKey }).map(line => (
                            <path
                                key={`stack-${line.dataKey as string}`}
                                d={line.path}
                                stroke="transparent"
                                opacity={0.5}
                                fill={line.color}
                            />
                        ))}
                    </Group>
                )}

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
                <Tooltip>
                    <TooltipContent series={series} xAxisKey={xAxisKey} activePoint={activePoint} />
                </Tooltip>
            )}
        </svg>
    )
}
