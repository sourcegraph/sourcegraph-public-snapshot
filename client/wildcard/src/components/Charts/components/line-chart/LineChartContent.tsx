import { type CSSProperties, type ReactElement, type SVGProps, useContext, useEffect, useMemo, useState } from 'react'

import { Group } from '@visx/group'
import { voronoi } from '@visx/voronoi'
import type { ScaleLinear, ScaleTime } from 'd3-scale'

import { SVGRootContext } from '../../core'
import type { Series } from '../../types'

import { LineDataSeries, StackedArea, Tooltip, TooltipContent } from './components'
import { getClosesVoronoiPoint, isNextTargetWithinCurrent } from './event-helpers'
import { useKeyboardNavigation } from './keyboard-navigation'
import type { Point } from './types'
import { generatePointsField, type SeriesWithData } from './utils'

import styles from './LineChartContent.module.scss'

interface GetLineGroupStyleOptions {
    /** The id of the series */
    id: string

    /** Whether this series contains the active point */
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
    onDatumClick: (point: Point) => void
    onDatumAreaClick: (point: Point) => void
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
        onDatumAreaClick,
        getActiveSeries,
        getLineGroupStyle,
        ...attributes
    } = props

    const { svgElement } = useContext(SVGRootContext)
    const [activePoint, setActivePoint] = useState<Point>()

    const activeSeries = useMemo(() => getActiveSeries(dataSeries), [getActiveSeries, dataSeries])
    const currentSeries = useMemo(
        () => activeSeries.find(series => series.id === activeSeriesId || series.id === activePoint?.seriesId),
        [activeSeries, activeSeriesId, activePoint?.seriesId]
    )

    const voronoiLayout = useMemo(() => {
        const points = generatePointsField(activeSeries)

        return voronoi<Point>({
            x: point => xScale(point.xValue),
            y: point => yScale(point.yValue),
            width,
            height,
        })(Object.values(points).flat())
    }, [activeSeries, height, width, xScale, yScale])

    // Experimental grid-like keyboard navigation over chart data points
    useKeyboardNavigation({ element: svgElement, series: activeSeries })

    // Listen all pointer events on the SVG element level in order to make
    // all chart surface interactive.
    useEffect(() => {
        if (!svgElement) {
            return
        }

        // Activate/deactivate cursor-like styles for all SVG root element
        syncCursorStyle(svgElement, !!activePoint?.linkUrl)

        const handlePointerMove = (event: PointerEvent): void => {
            const closestPoint = getClosesVoronoiPoint(event, voronoiLayout, { top, left })
            if (closestPoint && closestPoint.data.id !== activePoint?.id) {
                const element = svgElement.querySelector<Element>(`[data-id="${closestPoint.data.id}"]`)

                setActivePoint({ ...closestPoint.data, node: element ?? undefined })
            }
        }

        const handleMouseOrFocusOut = (event: PointerEvent | FocusEvent): void => {
            if (!isNextTargetWithinCurrent(event)) {
                setActivePoint(undefined)
            }
        }

        const handleClick = (): void => {
            if (activePoint) {
                onDatumAreaClick(activePoint)
            }
        }

        svgElement.addEventListener('pointermove', handlePointerMove, true)
        svgElement.addEventListener('pointerleave', handleMouseOrFocusOut)
        svgElement.addEventListener('click', handleClick)
        svgElement.addEventListener('blur', handleMouseOrFocusOut, true)

        return () => {
            svgElement.removeEventListener('pointermove', handlePointerMove, true)
            svgElement.removeEventListener('pointerleave', handleMouseOrFocusOut)
            svgElement.removeEventListener('click', handleClick)
            svgElement.removeEventListener('blur', handleMouseOrFocusOut, true)
        }
    }, [activePoint, left, onDatumAreaClick, svgElement, top, voronoiLayout])

    return (
        <Group {...attributes} role="list" aria-label="Chart series">
            {stacked && <StackedArea dataSeries={activeSeries} xScale={xScale} yScale={yScale} />}

            {activeSeries.map((line, index) => (
                <LineDataSeries
                    key={line.id}
                    id={line.id.toString()}
                    seriesIndex={index}
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
                    seriesIndex={-1}
                    xScale={xScale}
                    yScale={yScale}
                    dataset={currentSeries.data}
                    color={currentSeries.color}
                    activePointId={activePoint?.id}
                    tabIndex={-1}
                    pointerEvents="none"
                    aria-hidden={true}
                />
            )}

            {activePoint && (
                <Tooltip activeElement={activePoint.node as HTMLElement}>
                    <TooltipContent series={activeSeries} activePoint={activePoint} stacked={stacked} />
                </Tooltip>
            )}
        </Group>
    )
}

/**
 * We need to have a control over svg root element styles in order to change
 * cursor styles when we hover a link point area. This helper controls svg
 * cursor CSS class based on area link (cursor) argument
 */
function syncCursorStyle(element: Element, cursor: boolean): void {
    if (cursor) {
        element.classList.add(styles.svgWithCursor)
    } else {
        element.classList.remove(styles.svgWithCursor)
    }
}
