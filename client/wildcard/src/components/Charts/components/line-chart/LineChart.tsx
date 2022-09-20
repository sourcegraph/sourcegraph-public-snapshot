import { ReactElement, useMemo, useState, SVGProps, CSSProperties, useRef } from 'react'

import { Group } from '@visx/group'
import { scaleTime, scaleLinear, getTicks } from '@visx/scale'
import { AnyD3Scale } from '@visx/scale/lib/types/Scale'
import { voronoi } from '@visx/voronoi'
import classNames from 'classnames'
import { noop } from 'lodash'

import { AxisLeft, AxisBottom } from '../../core'
import { formatDateTick } from '../../core/components/axis/tick-formatters'
import { Series, SeriesLikeChart } from '../../types'

import { Tooltip, TooltipContent, LineDataSeries, StackedArea } from './components'
import { useChartEventHandlers } from './hooks/event-listeners'
import { Point } from './types'
import { getSeriesData, generatePointsField, getChartContentSizes, getMinMaxBoundaries } from './utils'

import styles from './LineChart.module.scss'

interface GetLineGroupStyleOptions {
    /** Whether this series contains the active point */
    id: string

    /** The id of the series */
    hasActivePoint: boolean

    /** Whether the chart has some active point */
    isActive: boolean
}

export interface LineChartProps<Datum> extends SeriesLikeChart<Datum>, SVGProps<SVGSVGElement> {
    /** The width of the chart */
    width: number

    /** The height of the chart */
    height: number

    /** Whether to start Y axis at zero */
    zeroYAxisMin?: boolean

    activeSeriesId?: string

    /** Function to style a given line group */
    getLineGroupStyle?: (options: GetLineGroupStyleOptions) => CSSProperties

    /**
     * If provided, uses this to render lines on the chart instead of `series`
     *
     * @param dataSeries a SeriesWithData array containing the data to render
     * @returns a SeriesWithData array that has been filtered
     */
    getActiveSeries?: <S extends Pick<Series<Datum>, 'id'>>(dataSeries: S[]) => S[]
}

const identity = <T,>(argument: T): T => argument

interface GetScaleTicksInput {
    scale: AnyD3Scale
    space: number
    pixelsPerTick?: number
}

export function getXScaleTicks<T>(input: GetScaleTicksInput): T[] {
    const { scale, space, pixelsPerTick = 80 } = input

    // Calculate desirable number of ticks
    const numberTicks = Math.max(2, Math.floor(space / pixelsPerTick))

    return getTicks(scale, numberTicks) as T[]
}

/**
 * Visual component that renders svg line chart with pre-defined sizes, tooltip,
 * voronoi area distribution.
 */
export function LineChart<D>(props: LineChartProps<D>): ReactElement | null {
    const {
        width: outerWidth,
        height: outerHeight,
        series,
        stacked = false,
        zeroYAxisMin = false,
        className,
        activeSeriesId,
        getLineGroupStyle,
        getActiveSeries = identity,
        onDatumClick = noop,
        ...attributes
    } = props

    const rootReference = useRef<SVGSVGElement>(null)
    const [activePoint, setActivePoint] = useState<Point>()
    const [yAxisElement, setYAxisElement] = useState<SVGGElement | null>(null)
    const [xAxisReference, setXAxisElement] = useState<SVGGElement | null>(null)

    const content = useMemo(
        () =>
            getChartContentSizes({
                width: outerWidth,
                height: outerHeight,
                margin: {
                    top: 16,
                    right: 16,
                    left: yAxisElement?.getBBox?.().width ?? 0,
                    bottom: xAxisReference?.getBBox?.().height ?? 0,
                },
            }),
        [yAxisElement, xAxisReference, outerWidth, outerHeight]
    )

    const dataSeries = useMemo(() => getSeriesData({ series, stacked }), [series, stacked])
    const activeSeries = useMemo(() => getActiveSeries(dataSeries), [getActiveSeries, dataSeries])

    const { minX, maxX, minY, maxY } = useMemo(() => getMinMaxBoundaries({ dataSeries, zeroYAxisMin }), [
        dataSeries,
        zeroYAxisMin,
    ])

    const xScale = useMemo(
        () =>
            scaleTime({
                domain: [minX, maxX],
                range: [0, content.width],
                nice: true,
                clamp: true,
            }),
        [minX, maxX, content]
    )

    const yScale = useMemo(
        () =>
            scaleLinear({
                domain: [minY, maxY],
                range: [content.height, 0],
                nice: true,
                clamp: true,
            }),
        [minY, maxY, content]
    )

    const voronoiLayout = useMemo(() => {
        const points = generatePointsField(activeSeries)

        return voronoi<Point>({
            // Taking into account content area shift in point distribution map
            // see https://github.com/sourcegraph/sourcegraph/issues/38919
            x: point => xScale(point.xValue) + content.left,
            y: point => yScale(point.yValue) + content.top,
            width: outerWidth,
            height: outerHeight,
        })(Object.values(points).flat())
    }, [activeSeries, outerWidth, outerHeight, xScale, content.left, content.top, yScale])

    const handlers = useChartEventHandlers({
        onPointerMove: point => {
            const closestPoint = voronoiLayout.find(point.x, point.y)

            if (closestPoint && closestPoint.data.id !== activePoint?.id) {
                setActivePoint(closestPoint.data)
            }
        },
        onClick: event => {
            if (activePoint?.linkUrl) {
                onDatumClick(event)
                window.open(activePoint.linkUrl)
            }
        },
        onPointerLeave: () => setActivePoint(undefined),
        onFocusOut: () => setActivePoint(undefined),
    })

    const currentSeries = useMemo(
        () => activeSeries.find(series => series.id === activeSeriesId || series.id === activePoint?.seriesId),
        [activeSeries, activeSeriesId, activePoint?.seriesId]
    )

    return (
        <svg
            ref={rootReference}
            width={outerWidth}
            height={outerHeight}
            className={classNames(styles.root, className, { [styles.rootWithHoveredLinkPoint]: activePoint?.linkUrl })}
            role="group"
            {...attributes}
            {...handlers}
        >
            <AxisLeft
                ref={setYAxisElement}
                scale={yScale}
                width={content.width}
                height={content.height}
                top={content.top}
                left={content.left}
            />

            <AxisBottom
                ref={setXAxisElement}
                scale={xScale}
                width={content.width}
                top={content.bottom}
                left={content.left}
                tickValues={getXScaleTicks({ scale: xScale, space: content.width })}
                tickFormat={formatDateTick}
            />

            <Group top={content.top} left={content.left} role="list">
                {stacked && <StackedArea dataSeries={activeSeries} xScale={xScale} yScale={yScale} />}

                {activeSeries.map(line => (
                    <LineDataSeries
                        key={line.id}
                        id={line.id.toString()}
                        xScale={xScale}
                        yScale={yScale}
                        dataset={line.data}
                        color={line.color}
                        getLinkURL={line.getLinkURL}
                        style={getLineGroupStyle?.({
                            id: `${line.id}`,
                            hasActivePoint: !!activePoint,
                            isActive: activePoint?.seriesId === line.id,
                        })}
                        aria-label={line.name}
                        role="listitem"
                        onDatumClick={onDatumClick}
                        onDatumFocus={setActivePoint}
                    />
                ))}

                {currentSeries && (
                    // Render line chart on top of all other lines and points in order to
                    // solve problem with z-index for SVG elements.
                    <LineDataSeries
                        id={currentSeries.id.toString()}
                        xScale={xScale}
                        yScale={yScale}
                        dataset={currentSeries.data}
                        color={currentSeries.color}
                        activePointId={activePoint?.id}
                        getLinkURL={currentSeries.getLinkURL}
                        onDatumClick={onDatumClick}
                        onDatumFocus={setActivePoint}
                        tabIndex={-1}
                        pointerEvents="none"
                        aria-hidden={true}
                    />
                )}
            </Group>

            {activePoint && rootReference.current && (
                <Tooltip containerElement={rootReference.current} activeElement={activePoint.node as HTMLElement}>
                    <TooltipContent series={activeSeries} activePoint={activePoint} stacked={stacked} />
                </Tooltip>
            )}
        </svg>
    )
}
