import { ReactElement, useMemo, useState, SVGProps, CSSProperties, useRef } from 'react'

import { Group } from '@visx/group'
import { scaleTime, scaleLinear, getTicks } from '@visx/scale'
import { AnyD3Scale } from '@visx/scale/lib/types/Scale'
import { LinePath } from '@visx/shape'
import { voronoi } from '@visx/voronoi'
import classNames from 'classnames'
import { ScaleLinear, ScaleTime } from 'd3-scale';
import { noop } from 'lodash'

import { AxisLeft, AxisBottom } from '../../core'
import { formatDateTick } from '../../core/components/axis/tick-formatters'
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
} from './utils'
import { GeneratedPoint } from './utils/generate-points-field';

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
    getActiveSeries?: <D>(dataSeries: SeriesWithData<D>[]) => SeriesWithData<D>[]
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
            voronoi<GeneratedPoint>({
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

    const currentSeries = useMemo(
        () => activeSeries.find(series => series.id === activeSeriesId || series.id === activePoint?.seriesId),
        [activeSeries, activeSeriesId, activePoint?.seriesId]
    )

    console.log({ activeSeriesId, activeSeries })

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
                tickValues={getXScaleTicks({ scale: xScale, space: content.width })}
                tickFormat={formatDateTick}
            />

            <Group top={content.top} left={content.left}>
                {stacked && <StackedArea dataSeries={activeSeries} xScale={xScale} yScale={yScale} />}

                {activeSeries.map(line => (
                    <DataSeries
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

                        onDatumClick={onDatumClick}
                        onDatumFocus={setActivePoint}/>
                ))}

                {currentSeries &&
                    <DataSeries
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
                        pointerEvents='none'
                        aria-hidden={true}
                    />
                }
            </Group>

            {activePoint && rootReference.current && (
                <Tooltip containerElement={rootReference.current} activeElement={activePoint.node as HTMLElement}>
                    <TooltipContent series={activeSeries} activePoint={activePoint} stacked={stacked} />
                </Tooltip>
            )}
        </svg>
    )
}

interface DataSeriesProps<D> extends SVGProps<SVGGElement> {
    id: string
    xScale: ScaleTime<number, number>
    yScale: ScaleLinear<number, number>
    dataset: SeriesDatum<D>[]
    color: string | undefined
    activePointId?: string
    getLinkURL?: (datum: D, index: number) => string | undefined
    onDatumClick: () => void
    onDatumFocus: (point: Point) => void
}

const NULL_LINK = (): undefined => undefined

function DataSeries<D>(props: DataSeriesProps<D>): ReactElement {
    const {
        id,
        xScale,
        yScale,
        dataset,
        color = 'green',
        activePointId,
        tabIndex,
        getLinkURL = NULL_LINK,
        onDatumClick,
        onDatumFocus,
        ...attributes
    } = props

    return (
        <Group tabIndex={tabIndex} {...attributes}>
            <LinePath
                data={dataset}
                defined={isDatumWithValidNumber}
                x={data => xScale(data.x)}
                y={data => yScale(getDatumValue(data))}
                stroke={color}
                strokeLinecap="round"
                strokeWidth={2}
            />

            { dataset.map((datum, index) => {
                const datumValue = getDatumValue(datum)
                const link = getLinkURL(datum.datum, index)
                const pointId = `${id}-${index}`

                return (
                    <PointGlyph
                        key={pointId}
                        tabIndex={tabIndex}
                        top={yScale(datumValue)}
                        left={xScale(datum.x)}
                        active={activePointId === pointId}
                        color={color}
                        linkURL={link}
                        onClick={onDatumClick}
                        onFocus={event => onDatumFocus({
                            id: pointId,
                            xValue: datum.x,
                            yValue: datumValue,
                            seriesId: id,
                            linkUrl: link,
                            node: event.target,
                        })}
                    />
                )
            })}
        </Group>
    )
}
