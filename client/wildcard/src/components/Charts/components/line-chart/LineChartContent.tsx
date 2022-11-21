import { CSSProperties, ReactElement, SVGProps, useMemo, useRef, useState } from 'react'

import { Group } from '@visx/group'
import { voronoi } from '@visx/voronoi'
import classNames from 'classnames'
import { ScaleLinear, ScaleTime } from 'd3-scale'

import { Series } from '../../types'

import { LineDataSeries, StackedArea, Tooltip, TooltipContent } from './components'
import { useChartEventHandlers } from './hooks/event-listeners'
import { Point } from './types'
import { generatePointsField, SeriesWithData } from './utils'

import styles from './LineChartContent.module.scss'

interface GetLineGroupStyleOptions {
    /** Whether this series contains the active point */
    id: string

    /** The id of the series */
    hasActivePoint: boolean

    /** Whether the chart has some active point */
    isActive: boolean
}

interface LineChartContentProps<Datum> extends SVGProps<SVGGElement> {
    width: number
    height: number
    top: number
    left: number
    stacked: boolean
    xScale: ScaleTime<number, number>
    yScale: ScaleLinear<number, number>

    dataSeries: SeriesWithData<Datum>[]
    activeSeriesId?: string
    getActiveSeries: <S extends Pick<Series<Datum>, 'id'>>(dataSeries: S[]) => S[]
    getLineGroupStyle?: (options: GetLineGroupStyleOptions) => CSSProperties
    onDatumClick: () => void
}

export function LineChartContent<Datum>(props: LineChartContentProps<Datum>): ReactElement {
    const {
        width,
        height,
        left,
        top,
        stacked,
        xScale,
        yScale,
        dataSeries,
        activeSeriesId,
        onDatumClick,
        getActiveSeries,
        getLineGroupStyle,
        className,
        ...attributes
    } = props

    const rootRef = useRef<SVGGElement>(null)
    const [activePoint, setActivePoint] = useState<Point>()

    const activeSeries = useMemo(() => getActiveSeries(dataSeries), [getActiveSeries, dataSeries])
    const currentSeries = useMemo(
        () => activeSeries.find(series => series.id === activeSeriesId || series.id === activePoint?.seriesId),
        [activeSeries, activeSeriesId, activePoint?.seriesId]
    )

    const voronoiLayout = useMemo(() => {
        const points = generatePointsField(activeSeries)

        return voronoi<Point>({
            // Taking into account content area shift in point distribution map
            // see https://github.com/sourcegraph/sourcegraph/issues/38919
            x: point => xScale(point.xValue),
            y: point => yScale(point.yValue),
            width,
            height,
        })(Object.values(points).flat())
    }, [activeSeries, height, width, xScale, yScale])

    const { onPointerLeave, onPointerMove, onClick, onBlurCapture } = useChartEventHandlers({
        onPointerMove: point => {
            const closestPoint = voronoiLayout.find(point.x - left, point.y - top)

            if (closestPoint && closestPoint.data.id !== activePoint?.id) {
                setActivePoint(closestPoint.data)
            }
        },
        onClick: () => {
            if (activePoint?.linkUrl) {
                onDatumClick()
                window.open(activePoint.linkUrl)
            }
        },
        onPointerLeave: () => setActivePoint(undefined),
        onFocusOut: () => setActivePoint(undefined),
    })

    return (
        <Group
            innerRef={rootRef}
            role="list"
            aria-label="Chart series"
            className={classNames(styles.root, className, { [styles.rootWithHoveredLinkPoint]: activePoint?.linkUrl })}
            onBlurCapture={onBlurCapture}
            {...attributes}
        >
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
                    aria-label={line.name}
                    role="listitem"
                    style={getLineGroupStyle?.({
                        id: `${line.id}`,
                        hasActivePoint: !!activePoint,
                        isActive: activePoint?.seriesId === line.id,
                    })}
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

            {/* Spread the group element in order to track user input events from all visible chart content surface*/}
            <rect
                x={0}
                y={0}
                opacity={0}
                width={width}
                height={height}
                aria-hidden={true}
                onPointerLeave={onPointerLeave}
                onPointerMove={onPointerMove}
                onClick={onClick}
            />

            {activePoint && rootRef.current && (
                <Tooltip containerElement={rootRef.current} activeElement={activePoint.node as HTMLElement}>
                    <TooltipContent series={activeSeries} activePoint={activePoint} stacked={stacked} />
                </Tooltip>
            )}
        </Group>
    )
}
