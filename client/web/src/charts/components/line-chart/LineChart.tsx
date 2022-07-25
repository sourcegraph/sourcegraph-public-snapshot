import { ReactElement, useMemo, useState, SVGProps, CSSProperties, useRef } from 'react'

import { AxisScale, TickFormatter } from '@visx/axis/lib/types'
import { Group } from '@visx/group'
import { scaleTime, scaleLinear } from '@visx/scale'
import { LinePath } from '@visx/shape'
import { voronoi } from '@visx/voronoi'
import classNames from 'classnames'
import { noop } from 'lodash'

import { AxisLeft, AxisBottom } from '../../core'
import { SeriesLikeChart } from '../../types'

import { Tooltip, TooltipContent, PointGlyph } from './components'
import { StackedArea } from './components/stacked-area/StackedArea'
import { useChartEventHandlers } from './hooks/event-listeners'
import { Point } from './types'
import {
    SeriesDatum,
    getDatumValue,
    isDatumWithValidNumber,
    getSeriesData,
    generatePointsField,
    getChartContentSizes,
    getMinMaxBoundaries,
    SeriesWithData,
    formatXTick,
} from './utils'

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

    /** Function to style a given line group */
    getLineGroupStyle?: (options: GetLineGroupStyleOptions) => CSSProperties

    /**
     * If provided, uses this to render lines on the chart instead of `series`
     *
     * @param dataSeries a SeriesWithData array containing the data to render
     * @returns a SeriesWithData array that has been filtered
     */
    getActiveSeries?: <D>(dataSeries: SeriesWithData<D>[]) => SeriesWithData<D>[]
}

const sortByDataKey = (dataKey: string | number | symbol, activeDataKey: string): number =>
    dataKey === activeDataKey ? 1 : -1

const identity = <T,>(argument: T): T => argument

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
        getLineGroupStyle,
        getActiveSeries = identity,
        onDatumClick = noop,
        ...attributes
    } = props

    const rootReference = useRef<SVGSVGElement>(null)
    const [activePoint, setActivePoint] = useState<Point<D> & { element?: Element }>()
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
                    left: yAxisElement?.getBBox().width,
                    bottom: xAxisReference?.getBBox().height,
                },
            }),
        [yAxisElement, xAxisReference, outerWidth, outerHeight]
    )

    const dataSeries = useMemo(() => getSeriesData({ series, stacked }), [series, stacked])
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

    const activeSeries = useMemo(() => getActiveSeries(dataSeries), [getActiveSeries, dataSeries])
    const points = useMemo(
        () =>
            generatePointsField({
                dataSeries: activeSeries,
                yScale,
                xScale,
            }),
        [activeSeries, yScale, xScale]
    )

    const voronoiLayout = useMemo(
        () =>
            voronoi<Point<D>>({
                // Taking into account content area shift in point distribution map
                // see https://github.com/sourcegraph/sourcegraph/issues/38919
                x: point => point.x + content.left,
                y: point => point.y + content.top,
                width: outerWidth,
                height: outerHeight,
            })(Object.values(points).flat()),
        [content, outerWidth, outerHeight, points]
    )

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

    const activeSeriesId = activePoint?.seriesId ?? ''
    const sortedSeries = useMemo(
        () =>
            [...activeSeries]
                // resorts array based on hover state
                // this is to make sure the hovered series is always rendered on top
                // since SVGs do not support z-index, we have to render the hovered
                // series last
                .sort(series => sortByDataKey(series.id, activeSeriesId)),
        [activeSeries, activeSeriesId]
    )

    return (
        <svg
            ref={rootReference}
            width={outerWidth}
            height={outerHeight}
            className={classNames(styles.root, className, { [styles.rootWithHoveredLinkPoint]: activePoint?.linkUrl })}
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
                tickFormat={(formatXTick as unknown) as TickFormatter<AxisScale>}
            />

            <Group top={content.top} left={content.left}>
                {stacked && <StackedArea dataSeries={activeSeries} xScale={xScale} yScale={yScale} />}

                {sortedSeries.map(line => (
                    <LinePath
                        key={line.id}
                        data={line.data as SeriesDatum<D>[]}
                        defined={isDatumWithValidNumber}
                        x={data => xScale(data.x)}
                        y={data => yScale(getDatumValue(data))}
                        stroke={line.color}
                        strokeLinecap="round"
                        strokeWidth={2}
                    />
                ))}

                {activeSeries.map(line => (
                    <Group
                        key={line.id}
                        style={getLineGroupStyle?.({
                            id: `${line.id}`,
                            hasActivePoint: !!activePoint,
                            isActive: activePoint?.seriesId === line.id,
                        })}
                    >
                        {points[line.id].map(point => (
                            <PointGlyph
                                key={point.id}
                                left={point.x}
                                top={point.y}
                                active={false}
                                color={point.color}
                                linkURL={point.linkUrl}
                                onClick={onDatumClick}
                                onFocus={event => setActivePoint({ ...point, element: event.target })}
                            />
                        ))}
                    </Group>
                ))}

                {activePoint && (
                    <PointGlyph
                        left={activePoint.x}
                        top={activePoint.y}
                        active={true}
                        color={activePoint.color}
                        linkURL={activePoint.linkUrl}
                        onClick={onDatumClick}
                        tabIndex={-1}
                    />
                )}
            </Group>

            {activePoint && rootReference.current && (
                <Tooltip containerElement={rootReference.current} activeElement={activePoint.element as HTMLElement}>
                    <TooltipContent series={activeSeries} activePoint={activePoint} stacked={stacked} />
                </Tooltip>
            )}
        </svg>
    )
}
