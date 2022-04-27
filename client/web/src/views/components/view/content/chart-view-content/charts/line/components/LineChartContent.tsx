import React, { ReactElement, useCallback, useMemo, useState, MouseEvent, useRef } from 'react'

import { curveLinear } from '@visx/curve'
import { GridRows } from '@visx/grid'
import { Group } from '@visx/group'
import {
    Axis,
    DataProvider,
    GlyphSeries,
    LineSeries,
    Tooltip,
    TooltipProvider,
    XYChart,
    EventEmitterProvider,
} from '@visx/xychart'
import { RenderTooltipParams } from '@visx/xychart/lib/components/Tooltip'
import { XYCHART_EVENT_SOURCE } from '@visx/xychart/lib/constants'
import isValidNumber from '@visx/xychart/lib/typeguards/isValidNumber'
import { EventHandlerParams } from '@visx/xychart/lib/types'
import classNames from 'classnames'
import { noop } from 'rxjs'
import { LineChartContent as LineChartContentType, LineChartSeries } from 'sourcegraph'

import { DEFAULT_LINE_STROKE } from '../constants'
import { generateAccessors } from '../helpers/generate-accessors'
import { getProcessedChartData } from '../helpers/get-processed-chart-data'
import { getYAxisWidth } from '../helpers/get-y-axis-width'
import { getYTicks } from '../helpers/get-y-ticks'
import { usePointerEventEmitters } from '../helpers/use-event-emitters'
import { useScalesConfiguration, useXScale, useYScale } from '../helpers/use-scales'
import { onDatumZoneClick, Point } from '../types'

import { ActiveDatum, GlyphContent } from './GlyphContent'
import { NonActiveBackground } from './NonActiveBackground'
import { dateTickFormatter, numberFormatter, Tick, getTickXProps, getTickYProps } from './TickComponent'
import { TooltipContent } from './tooltip-content/TooltipContent'

import styles from './LineChartContent.module.scss'

// Chart configuration
const WIDTH_PER_TICK = 70
const MARGIN = { top: 10, left: 0, bottom: 26, right: 20 }
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
export const getLineStroke = <Datum extends object>(line: LineChartSeries<Datum>): string =>
    line?.stroke ?? DEFAULT_LINE_STROKE

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

    /**
     * Callback calls every time when link-point and only link-point
     * on the chart was clicked.
     */
    onDatumLinkClick?: (event: React.MouseEvent) => void
}

export function LineChart<Datum extends object>(props: LineChartContentProps<Datum>): ReactElement {
    return (
        <EventEmitterProvider>
            <LineChartContent {...props} />
        </EventEmitterProvider>
    )
}

/**
 * Displays line chart content - line chart, tooltip, active point
 */
export function LineChartContent<Datum extends object>(props: LineChartContentProps<Datum>): ReactElement {
    const { width, height, data, series, xAxis, onDatumZoneClick = noop, onDatumLinkClick = noop } = props

    // XYChart must know how to get the right data from datum object in order to render lines and axes
    // Because of that we have to generate map of getters for all kind of data which will be rendered on the chart.
    const accessors = useMemo(() => generateAccessors(xAxis, series), [xAxis, series])

    const scalesConfiguration = useScalesConfiguration({
        data,
        accessors,
        config: SCALES_CONFIG,
    })

    const innerHeight = height - MARGIN.top - MARGIN.bottom

    const yScale = useYScale({ config: scalesConfiguration.y, height: innerHeight })
    const yTicks = getYTicks(yScale, innerHeight)
    const yAxisWidth = getYAxisWidth(yTicks)

    // Calculate inner sizes for chart without padding values
    const innerWidth = width - MARGIN.left - MARGIN.right - yAxisWidth
    const numberOfTicksX = Math.max(1, Math.floor(innerWidth / WIDTH_PER_TICK))

    const xScale = useXScale({
        config: scalesConfiguration.x,
        width: innerWidth,
        accessors,
        data,
    })

    const dynamicMargin = { ...MARGIN, left: MARGIN.left + yAxisWidth }

    const { sortedData, seriesWithData } = useMemo(() => getProcessedChartData({ accessors, data, series }), [
        data,
        accessors,
        series,
    ])

    // state
    const [hoveredDatum, setHoveredDatum] = useState<ActiveDatum<Datum> | null>(null)
    const [focusedDatum, setFocusedDatum] = useState<ActiveDatum<Datum> | null>(null)

    // callbacks
    const renderTooltip = useCallback(
        (renderProps: RenderTooltipParams<Point>) => <TooltipContent {...renderProps} series={seriesWithData} />,
        [seriesWithData]
    )

    const handlePointerMove = useCallback(
        (event: EventHandlerParams<Point>) => {
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
        (info: EventHandlerParams<Point>) => {
            info.event?.persist()

            // According to types from visx/xychart index can be undefined
            const activeDatumIndex = hoveredDatum?.index
            const line = series.find(line => line.dataKey === info.key)

            if (!info.event || !line || !hoveredDatum?.datum || !isValidNumber(activeDatumIndex)) {
                return
            }

            onDatumZoneClick({
                originEvent: info.event as MouseEvent<unknown>,
                link: line?.linkURLs?.[+hoveredDatum.datum.x] ?? line?.linkURLs?.[activeDatumIndex],
            })
        },
        [series, onDatumZoneClick, hoveredDatum]
    )

    const { onPointerMove = noop, onPointerOut = noop, ...otherHandlers } = usePointerEventEmitters({
        source: XYCHART_EVENT_SOURCE,
        onFocus: true,
        onBlur: true,
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

    // Disable all event listeners explicitly to avoid flaky tooltip appearance
    const eventEmitters = {
        onPointerMove: handleRootPointerMove,
        onPointerOut: handleRootPointerOut,
        ...otherHandlers,
    }

    const hoveredDatumLinks = hoveredDatum?.line?.linkURLs ?? {}
    const hoveredDatumLink = hoveredDatum
        ? hoveredDatumLinks[+hoveredDatum.datum.x] ?? hoveredDatumLinks[hoveredDatum.index]
        : null
    const rootClasses = classNames({ [styles.contentWithCursor]: !!hoveredDatumLink })

    return (
        <div className={classNames(rootClasses, 'percy-inactive-element')} data-testid="line-chart__content">
            {/*
                Because XYChart wraps itself with context providers in case if consumer didn't add them
                But this recursive wrapping leads to problem with event emitter context - double subscription all event
                See https://github.com/airbnb/visx/blob/master/packages/visx-xychart/src/components/XYChart.tsx#L128-L138
                If we need override EventEmitter (our case because we have to capture all event by ourselves) we
                have to provide DataContext and TooltipContext as well to avoid problem with EmitterContext.
            */}
            <DataProvider
                xScale={scalesConfiguration.x}
                yScale={scalesConfiguration.y}
                initialDimensions={{ width, height, margin: dynamicMargin }}
            >
                <TooltipProvider hideTooltipDebounceMs={0}>
                    <XYChart
                        height={height}
                        width={width}
                        captureEvents={false}
                        margin={dynamicMargin}
                        onPointerMove={handlePointerMove}
                        onPointerUp={handlePointerUp}
                        accessibilityLabel="Line chart content"
                    >
                        <NonActiveBackground data={sortedData} accessors={accessors} series={series} />
                        <Group aria-label="Chart axes">
                            {/* eslint-disable-next-line jsx-a11y/aria-role */}
                            <Group role="graphics-axis" aria-orientation="horizontal" aria-label="Y axis: number">
                                <Group aria-hidden={true} top={dynamicMargin.top} left={dynamicMargin.left}>
                                    <GridRows
                                        scale={yScale}
                                        tickValues={yTicks}
                                        width={innerWidth}
                                        className={styles.gridLine}
                                    />
                                </Group>

                                <Axis
                                    orientation="left"
                                    tickValues={yTicks}
                                    tickFormat={numberFormatter}
                                    /* eslint-disable-next-line @typescript-eslint/ban-ts-comment */
                                    // @ts-ignore
                                    tickLabelProps={getTickYProps}
                                    tickComponent={Tick}
                                    axisLineClassName={classNames(styles.axisLine, styles.axisLineVertical)}
                                    tickClassName={classNames(styles.axisTick, styles.axisTickVertical)}
                                />
                            </Group>

                            {/* eslint-disable-next-line jsx-a11y/aria-role */}
                            <Group role="graphics-axis" aria-orientation="horizontal" aria-label="X axis: time">
                                <Axis
                                    orientation="bottom"
                                    tickValues={xScale.ticks(numberOfTicksX)}
                                    tickFormat={dateTickFormatter}
                                    numTicks={numberOfTicksX}
                                    /* eslint-disable-next-line @typescript-eslint/ban-ts-comment */
                                    // @ts-ignore
                                    tickLabelProps={getTickXProps}
                                    tickComponent={Tick}
                                    tickLength={8}
                                    axisLineClassName={styles.axisLine}
                                    tickClassName={styles.axisTick}
                                />
                            </Group>
                        </Group>

                        <Group
                            // eslint-disable-next-line jsx-a11y/aria-role
                            role="graphics-datagroup"
                            aria-label="Chart series"
                            pointerEvents="bounding-box"
                            // Check percy test run to be able disable flaky line chart tooltip appearance
                            // by disabling any point events over line chart container.
                            // See https://github.com/sourcegraph/sourcegraph/issues/23669
                            className="percy-inactive-element"
                            {...eventEmitters}
                        >
                            {/* Spread size of parent group element by transparent rect with width and height */}
                            <rect
                                x={dynamicMargin.left}
                                y={dynamicMargin.top}
                                width={innerWidth}
                                height={innerHeight}
                                aria-hidden={true}
                                fill="transparent"
                            />

                            {seriesWithData.map((line, index) => (
                                <Group
                                    key={line.dataKey as string}
                                    // eslint-disable-next-line jsx-a11y/aria-role
                                    role="graphics-datagroup"
                                    data-line-name={line.name ?? 'unknown'}
                                    aria-label={`Line ${index + 1} of ${series.length}. Name: ${
                                        line.name ?? 'unknown'
                                    }`}
                                >
                                    <LineSeries
                                        dataKey={line.dataKey as string}
                                        data={line.data}
                                        strokeWidth={2}
                                        /* eslint-disable-next-line jsx-a11y/aria-role */
                                        role="graphics-dataline"
                                        xAccessor={point => point?.x}
                                        yAccessor={point => point?.y}
                                        stroke={getLineStroke(line)}
                                        curve={curveLinear}
                                        aria-hidden={true}
                                    />

                                    <GlyphSeries
                                        dataKey={line.dataKey as string}
                                        data={line.data}
                                        enableEvents={false}
                                        xAccessor={point => point?.x}
                                        yAccessor={point => point?.y}
                                        // Don't have info about line in props. @visx/xychart doesn't expose this information
                                        // Move this arrow function in separate component when API of GlyphSeries will be fixed.
                                        renderGlyph={glyphProps => (
                                            <GlyphContent
                                                {...glyphProps}
                                                index={glyphProps.key}
                                                hoveredDatum={hoveredDatum}
                                                focusedDatum={focusedDatum}
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
                            className={styles.tooltip}
                            showHorizontalCrosshair={false}
                            showVerticalCrosshair={true}
                            snapTooltipToDatumX={false}
                            snapTooltipToDatumY={false}
                            showDatumGlyph={false}
                            showSeriesGlyphs={false}
                            verticalCrosshairStyle={{ strokeWidth: 2, stroke: 'var(--secondary)' }}
                            renderTooltip={renderTooltip}
                        />
                    </XYChart>
                </TooltipProvider>
            </DataProvider>
        </div>
    )
}
