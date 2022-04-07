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
import { StackedArea } from './components/stacked-area/StackedArea'
import { useChartEventHandlers } from './hooks/event-listeners'
import { Point } from './types'
import {
    getDatumValue,
    isDatumWithValidNumber,
    getSeriesData,
    generatePointsField,
    getChartContentSizes,
    getMinMaxBoundaries,
} from './utils'
import { SeriesDatum } from './utils/data-series-processing/types'

import styles from './LineChart.module.scss'

export interface LineChartContentProps<Datum> extends SeriesLikeChart<Datum> {
    width: number
    height: number
    zeroYAxisMin?: boolean
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
        stacked = false,
        zeroYAxisMin = false,
        getXValue,
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

    const dataSeries = useMemo(() => getSeriesData({ data, series, stacked, getXValue }), [
        data,
        series,
        stacked,
        getXValue,
    ])

    const { minX, maxX, minY, maxY } = useMemo(() => getMinMaxBoundaries({ dataSeries, zeroYAxisMin }), [
        dataSeries,
        zeroYAxisMin,
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
                clamp: true,
            }),
        [minY, maxY, margin.top, height]
    )

    const points = useMemo(() => generatePointsField({ dataSeries, getXValue, yScale, xScale }), [
        dataSeries,
        getXValue,
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
                width={width}
                height={height}
                top={margin.top}
                left={margin.left}
                getXValue={getXValue}
                xScale={xScale}
            />

            <Group top={margin.top}>
                {stacked && <StackedArea dataSeries={dataSeries} xScale={xScale} yScale={yScale} />}

                {dataSeries.map(line => (
                    <LinePath
                        key={line.dataKey as string}
                        data={line.data as SeriesDatum<D>[]}
                        curve={curveLinear}
                        defined={isDatumWithValidNumber}
                        x={data => xScale(data.x)}
                        y={data => yScale(getDatumValue(data))}
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
                    <TooltipContent series={series} activePoint={activePoint} stacked={stacked} />
                </Tooltip>
            )}
        </svg>
    )
}
