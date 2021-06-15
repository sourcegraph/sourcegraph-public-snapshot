import { curveLinear } from '@visx/curve'
import { GridRows } from '@visx/grid'
import { GridScale } from '@visx/grid/lib/types'
import { Group } from '@visx/group'
import { Axis, DataProvider, GlyphSeries, LineSeries, Tooltip, TooltipProvider, XYChart } from '@visx/xychart'
import { RenderTooltipParams } from '@visx/xychart/lib/components/Tooltip'
import { XYCHART_EVENT_SOURCE } from '@visx/xychart/lib/constants'
import isValidNumber from '@visx/xychart/lib/typeguards/isValidNumber'
import { EventHandlerParams } from '@visx/xychart/lib/types'
import classnames from 'classnames'
import React, { ReactElement, useCallback, useMemo, useState, MouseEvent, useRef } from 'react'
import { noop } from 'rxjs'
import { LineChartContent as LineChartContentType } from 'sourcegraph'

import { DEFAULT_LINE_STROKE } from '../constants'
import { generateAccessors } from '../helpers/generate-accessors'
import { usePointerEventEmitters } from '../helpers/use-event-emitters'
import { useScales } from '../helpers/use-scales'
import { onDatumZoneClick } from '../types'

import { ActiveDatum, GlyphContent } from './GlyphContent'
import { NonActiveBackground } from './NonActiveBackground'
import { dateTickFormatter, numberFormatter, Tick, getTickXProps, getTickYProps } from './TickComponent'
import { TooltipContent } from './TooltipContent'

// Chart configuration
const WIDTH_PER_TICK = 70
const HEIGHT_PER_TICK = 40
const MARGIN = { top: 10, left: 30, bottom: 26, right: 20 }
const SCALES_CONFIG = {
    x: {
        type: 'time' as const,
        nice: true,
    },
    y: {
        type: 'linear' as const,
        nice: true,
        zero: false,
        clamp: false,
    },
}

// Line color accessor
export const getLineStroke = <Datum extends object>(
    line: LineChartContentType<Datum, keyof Datum>['series'][number]
): string => line?.stroke ?? DEFAULT_LINE_STROKE

const stopPropagation = (event: React.MouseEvent): void => event.stopPropagation()

export interface LineChartContentProps<Datum extends object>
    extends Omit<LineChartContentType<Datum, keyof Datum>, 'chart'> {
    /** Chart width value in px */
    width: number
    /** Chart height value in px */
    height: number
    /**
     * Callback calls every time when a point-zone (zone around point) but not point itself
     * on the chart was clicked.
     */
    onDatumZoneClick?: onDatumZoneClick
    /** Callback calls every time when link-point and only link-point on the chart was clicked. */
    onDatumLinkClick?: (event: React.MouseEvent) => void
}

/**
 * Displays line chart content - line chart, tooltip, active point
 * */
export function LineChartContent<Datum extends object>(props: LineChartContentProps<Datum>): ReactElement {
    const { width, height, data, series, xAxis, onDatumZoneClick = noop, onDatumLinkClick = noop } = props

    // Calculate inner sizes for chart without padding values
    const innerWidth = width - MARGIN.left - MARGIN.right
    const innerHeight = height - MARGIN.top - MARGIN.bottom

    // Calculate how many labels we need to have for each axis
    const numberOfTicksX = Math.max(1, Math.floor(innerWidth / WIDTH_PER_TICK))
    const numberOfTicksY = Math.max(1, Math.floor(innerHeight / HEIGHT_PER_TICK))

    // In case if we've got unsorted by x (time) axis dataset we have to sort that by ourselves
    // otherwise we will get an error in calculation of position for the tooltip
    // Details: bisector from d3-array package expects sorted data otherwise he can't calculate
    // right index for nearest point on the XYChart.
    // See https://github.com/airbnb/visx/blob/master/packages/visx-xychart/src/utils/findNearestDatumSingleDimension.ts#L30
    const sortedData = useMemo(
        () => data.sort((firstDatum, secondDatum) => +firstDatum[xAxis.dataKey] - +secondDatum[xAxis.dataKey]),
        [data, xAxis]
    )

    // XYChart must know how to get the right data from datum object in order to render lines and axes
    // Because of that we have to generate map of getters for all kind of data which will be rendered on the chart.
    const accessors = useMemo(() => generateAccessors(xAxis, series), [xAxis, series])

    const { config: scalesConfig, xScale, yScale } = useScales({
        config: SCALES_CONFIG,
        data: sortedData,
        width: innerWidth,
        height: innerHeight,
        accessors,
    })

    // state
    const [hoveredDatum, setHoveredDatum] = useState<ActiveDatum<Datum> | null>(null)
    const [focusedDatum, setFocusedDatum] = useState<ActiveDatum<Datum> | null>(null)

    // callbacks
    const renderTooltip = useCallback(
        (renderProps: RenderTooltipParams<Datum>) => (
            <TooltipContent
                {...renderProps}
                accessors={accessors}
                series={series}
                className="line-chart__tooltip-content"
            />
        ),
        [accessors, series]
    )

    const handlePointerMove = useCallback(
        (event: EventHandlerParams<Datum>) => {
            // If active point hasn't been change we shouldn't call setActiveDatum again
            if (hoveredDatum?.index === event.index && hoveredDatum?.key === event.key) {
                return
            }

            const line = series.find(line => line.dataKey === event.key)

            if (!line) {
                setHoveredDatum(null)
                return
            }

            setHoveredDatum({
                ...event,
                line,
            })
        },
        [series, hoveredDatum]
    )

    const handlePointerUp = useCallback(
        (info: EventHandlerParams<Datum>) => {
            info.event?.persist()

            // According to types from visx/xychart index can be undefined
            const activeDatumIndex = hoveredDatum?.index
            const line = series.find(line => line.dataKey === info.key)

            if (!info.event || !line || !isValidNumber(activeDatumIndex)) {
                return
            }

            onDatumZoneClick({
                originEvent: info.event as MouseEvent<unknown>,
                link: line?.linkURLs?.[activeDatumIndex],
            })
        },
        [series, onDatumZoneClick, hoveredDatum]
    )

    const { onPointerMove = noop, onPointerOut = noop, ...otherHandlers } = usePointerEventEmitters({
        source: XYCHART_EVENT_SOURCE,
    })

    // We only need to catch pointerout event on root element - chart
    // we can't rely on event propagation here because this leads us to
    // unnecessary calls when some child element had lost cursor he fired
    // that unnecessary event. So we have to track the pointerout by ourselves.
    // This focused ref is kind of a flag to track do we have any event from
    // user on chart or not used below in move and out handlers to fire pointerout
    // event in right moment and avoid unnecessary onPointerOut calls.
    const focused = useRef(false)

    const handleRootPointerMove = useCallback(
        (event: React.PointerEvent) => {
            // Track user activity over chart
            focused.current = true
            onPointerMove(event)
        },
        [onPointerMove]
    )

    const handleRootPointerOut = useCallback(
        (event: React.PointerEvent) => {
            event.persist()

            // Some element has lost cursor and fired pointerout event but
            // we don't know which element did that root element or some child element within root element
            // So we mark this focused state as false = root element is not active and then schedule
            // next frame check decide do we need fire callback or know. If child lost cursor then
            // but cursor still on chart then we mark this focused state in pointerMove handler above
            // and won't fire onPointerOut callback. In case if root element child lost cursor we fire onPointerOut
            focused.current = false

            requestAnimationFrame(() => {
                if (!focused.current) {
                    setHoveredDatum(null)
                    onPointerOut(event)
                }
            })
        },
        [focused, onPointerOut, setHoveredDatum]
    )

    const eventEmitters = {
        onPointerMove: handleRootPointerMove,
        onPointerOut: handleRootPointerOut,
        ...otherHandlers,
    }

    const hoveredDatumLink = hoveredDatum?.line?.linkURLs?.[hoveredDatum?.index]
    const rootClasses = classnames('line-chart__content', { 'line-chart__content--with-cursor': !!hoveredDatumLink })

    return (
        <div className={rootClasses} data-testid="line-chart__content">
            {/*
                Because XYChart wraps itself with context providers in case if consumer didn't add them
                But this recursive wrapping leads to problem with event emitter context - double subscription all event
                See https://github.com/airbnb/visx/blob/master/packages/visx-xychart/src/components/XYChart.tsx#L128-L138
                If we need override EventEmitter (our case because we have to capture all event by ourselves) we
                have to provide DataContext and TooltipContext as well to avoid problem with EmitterContext.
            */}
            <DataProvider
                xScale={scalesConfig.x}
                yScale={scalesConfig.y}
                initialDimensions={{ width, height, margin: MARGIN }}
            >
                <TooltipProvider>
                    <XYChart
                        height={height}
                        width={width}
                        captureEvents={false}
                        margin={MARGIN}
                        onPointerMove={handlePointerMove}
                        onPointerUp={handlePointerUp}
                        accessibilityLabel="Line chart content"
                    >
                        <NonActiveBackground data={sortedData} accessors={accessors} series={series} />
                        <Group aria-label="Chart axes">
                            {/* eslint-disable-next-line jsx-a11y/aria-role */}
                            <Group role="graphics-axis" aria-orientation="horizontal" aria-label="Y axis: number">
                                <Group aria-hidden={true} top={MARGIN.top} left={MARGIN.left}>
                                    <GridRows
                                        scale={yScale as GridScale}
                                        numTicks={numberOfTicksY}
                                        width={innerWidth}
                                        className="line-chart__grid-line"
                                    />
                                </Group>

                                <Axis
                                    orientation="left"
                                    numTicks={numberOfTicksY}
                                    tickFormat={numberFormatter}
                                    tickLabelProps={getTickYProps}
                                    tickComponent={Tick}
                                    axisClassName="line-chart__axis"
                                    axisLineClassName="line-chart__axis-line line-chart__axis-line--vertical"
                                    tickClassName="line-chart__axis-tick line-chart__axis-tick--vertical"
                                />
                            </Group>

                            {/* eslint-disable-next-line jsx-a11y/aria-role */}
                            <Group role="graphics-axis" aria-orientation="horizontal" aria-label="X axis: time">
                                <Axis
                                    orientation="bottom"
                                    tickValues={xScale.ticks(numberOfTicksX)}
                                    tickFormat={dateTickFormatter}
                                    numTicks={numberOfTicksX}
                                    tickLabelProps={getTickXProps}
                                    tickComponent={Tick}
                                    tickLength={8}
                                    axisClassName="line-chart__axis"
                                    axisLineClassName="line-chart__axis-line"
                                    tickClassName="line-chart__axis-tick"
                                />
                            </Group>
                        </Group>

                        <Group
                            // eslint-disable-next-line jsx-a11y/aria-role
                            role="graphics-datagroup"
                            aria-label="Chart series"
                            pointerEvents="bounding-box"
                            {...eventEmitters}
                        >
                            {/* Spread size of parent group element by transparent rect with width and height */}
                            <rect
                                x={MARGIN.left}
                                y={MARGIN.top}
                                width={innerWidth}
                                height={innerHeight}
                                aria-hidden={true}
                                fill="transparent"
                            />

                            {series.map((line, index) => (
                                <Group
                                    key={line.dataKey as string}
                                    // eslint-disable-next-line jsx-a11y/aria-role
                                    role="graphics-datagroup"
                                    aria-label={`Line ${index + 1} of ${series.length}. Name: ${
                                        line.name ?? 'unknown'
                                    }`}
                                >
                                    <LineSeries
                                        dataKey={line.dataKey as string}
                                        data={sortedData}
                                        strokeWidth={2}
                                        /* eslint-disable-next-line jsx-a11y/aria-role */
                                        role="graphics-dataline"
                                        xAccessor={accessors.x}
                                        yAccessor={accessors.y[line.dataKey]}
                                        stroke={getLineStroke(line)}
                                        curve={curveLinear}
                                        aria-hidden={true}
                                    />

                                    <GlyphSeries
                                        dataKey={line.dataKey as string}
                                        data={sortedData}
                                        enableEvents={false}
                                        xAccessor={accessors.x}
                                        yAccessor={accessors.y[line.dataKey]}
                                        // Don't have info about line in props. @visx/xychart doesn't expose this information
                                        // Move this arrow function in separate component when API of GlyphSeries will be fixed.
                                        renderGlyph={glyphProps => (
                                            <GlyphContent
                                                {...glyphProps}
                                                index={glyphProps.key}
                                                hoveredDatum={hoveredDatum}
                                                focusedDatum={focusedDatum}
                                                accessors={accessors}
                                                line={line}
                                                lineIndex={index}
                                                totalNumberOfLines={series.length}
                                                setFocusedDatum={setFocusedDatum}
                                                onPointerUp={stopPropagation}
                                                onClick={onDatumLinkClick}
                                            />
                                        )}
                                    />
                                </Group>
                            ))}
                        </Group>

                        <Tooltip
                            className="line-chart__tooltip"
                            showHorizontalCrosshair={false}
                            showVerticalCrosshair={true}
                            snapTooltipToDatumX={false}
                            snapTooltipToDatumY={false}
                            showDatumGlyph={false}
                            showSeriesGlyphs={false}
                            renderTooltip={renderTooltip}
                        />
                    </XYChart>
                </TooltipProvider>
            </DataProvider>
        </div>
    )
}
