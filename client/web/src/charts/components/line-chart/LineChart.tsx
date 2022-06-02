import { ReactElement, useMemo, useState, SVGProps, CSSProperties } from 'react'

import { Group } from '@visx/group'
import { scaleTime, scaleLinear } from '@visx/scale'
import { LinePath } from '@visx/shape'
import { voronoi } from '@visx/voronoi'
import classNames from 'classnames'
import { noop } from 'lodash'

import { SeriesLikeChart } from '../../types'

import { AxisBottom, AxisLeft, Tooltip, TooltipContent, PointGlyph } from './components'
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
} from './utils'

import styles from './LineChart.module.scss'

export interface LineChartContentProps<Datum> extends SeriesLikeChart<Datum>, SVGProps<SVGSVGElement> {
    width: number
    height: number
    zeroYAxisMin?: boolean
    isSeriesSelected: (id: string) => boolean
    isSeriesHovered: (id: string) => boolean
}

const sortByDataKey = (dataKey: string | number | symbol, activeDataKey: string): number =>
    dataKey === activeDataKey ? 1 : -1

/**
 * Visual component that renders svg line chart with pre-defined sizes, tooltip,
 * voronoi area distribution.
 */
export function LineChart<D>(props: LineChartContentProps<D>): ReactElement | null {
    const {
        width: outerWidth,
        height: outerHeight,
        series,
        stacked = false,
        zeroYAxisMin = false,
        onDatumClick = noop,
        className,
        isSeriesSelected,
        isSeriesHovered,
        ...attributes
    } = props

    const [activePoint, setActivePoint] = useState<Point<D> & { element?: Element }>()
    const [yAxisElement, setYAxisElement] = useState<SVGGElement | null>(null)
    const [xAxisReference, setXAxisElement] = useState<SVGGElement | null>(null)

    const { width, height, margin } = useMemo(
        () =>
            getChartContentSizes({
                width: outerWidth,
                height: outerHeight,
                margin: {
                    top: 10,
                    right: 20,
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
                range: [margin.left, outerWidth - margin.right],
                nice: true,
                clamp: true,
            }),
        [minX, maxX, margin.left, margin.right, outerWidth]
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

    const points = useMemo(() => generatePointsField({ dataSeries, yScale, xScale }), [dataSeries, yScale, xScale])

    const voronoiLayout = useMemo(
        () =>
            voronoi<Point<D>>({
                x: point => point.x,
                y: point => point.y,
                width: outerWidth,
                height: outerHeight,
            })(Object.values(points).flat()),
        [outerWidth, outerHeight, points]
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

    const getHoverStyle = (id: string): CSSProperties => {
        const opacity = isSeriesSelected(id) ? 1 : isSeriesHovered(id) ? 0.5 : 0

        return {
            opacity,
            transitionProperty: 'opacity',
            transitionDuration: '200ms',
            transitionTimingFunction: 'ease-out',
        }
    }

    return (
        <svg
            width={outerWidth}
            height={outerHeight}
            className={classNames(styles.root, className, { [styles.rootWithHoveredLinkPoint]: activePoint?.linkUrl })}
            {...attributes}
            {...handlers}
        >
            <AxisLeft
                ref={setYAxisElement}
                scale={yScale}
                width={width}
                height={height}
                top={margin.top}
                left={margin.left}
            />

            <AxisBottom ref={setXAxisElement} scale={xScale} top={margin.top + height} width={width} />

            <Group top={margin.top}>
                {stacked && <StackedArea dataSeries={dataSeries} xScale={xScale} yScale={yScale} />}

                {[...dataSeries]
                    .sort(series => sortByDataKey(series.id, activePoint?.seriesId || ''))
                    .map(line => (
                        <Group key={line.id} style={getHoverStyle(`${line.id}`)}>
                            <LinePath
                                data={line.data as SeriesDatum<D>[]}
                                defined={isDatumWithValidNumber}
                                x={data => xScale(data.x)}
                                y={data => yScale(getDatumValue(data))}
                                stroke={line.color}
                                strokeLinecap="round"
                                strokeWidth={2}
                            />
                            {points[line.id]
                                .sort(point => sortByDataKey(point.seriesId, activePoint?.seriesId || ''))
                                .map(point => (
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
