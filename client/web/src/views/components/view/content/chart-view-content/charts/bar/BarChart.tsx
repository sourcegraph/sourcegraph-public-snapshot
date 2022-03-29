import React, { ReactElement, useCallback, useMemo } from 'react'

import { AxisBottom, AxisLeft } from '@visx/axis'
import { localPoint } from '@visx/event'
import { GridRows } from '@visx/grid'
import { Group } from '@visx/group'
import { scaleBand, scaleLinear } from '@visx/scale'
import { Bar } from '@visx/shape'
import { useTooltip, TooltipWithBounds } from '@visx/tooltip'
import classNames from 'classnames'
import { range } from 'lodash'
import { BarChartContent } from 'sourcegraph'

import { LockedChart } from '../locked/LockedChart'
import { MaybeLink } from '../MaybeLink'

import styles from './BarChart.module.scss'

const DEFAULT_PADDING = { top: 20, right: 20, bottom: 25, left: 40 }

// Tooltip timeout used below as semaphore to prevent tooltip flashing
// in case if user is moving mouse very fast between bars
let tooltipTimeout: number

/** Data which needs to display tooltip with content. */
interface TooltipData {
    /** Label for current hovered bar */
    xLabel: string
    /** Y value for hovered bar */
    value: number
}

interface BarChartProps<Datum extends object> extends Omit<BarChartContent<Datum, keyof Datum>, 'chart'> {
    /** Chart width in px. */
    width: number
    /** Chart height in px. */
    height: number
    /** Callback calls every time when a bar-link on the chart was clicked */
    onDatumLinkClick: (event: React.MouseEvent) => void
    locked?: boolean
}

/**
 * Displays bar chart with tooltip.
 */
export function BarChart<Datum extends object>(props: BarChartProps<Datum>): ReactElement {
    const {
        width,
        height,
        data,
        series,
        onDatumLinkClick,
        xAxis: { dataKey: xDataKey },
        locked = false,
    } = props

    // Respect only first element of data series
    // Refactor this in case if we need support stacked bar chart
    const { dataKey, fill, linkURLs } = series[0]

    const innerWidth = width - DEFAULT_PADDING.left - DEFAULT_PADDING.right
    const innerHeight = height - DEFAULT_PADDING.top - DEFAULT_PADDING.bottom

    const { tooltipOpen, tooltipLeft, tooltipTop, tooltipData, hideTooltip, showTooltip } = useTooltip<TooltipData>()

    // Get access to y value of each bar (datum)
    const yAccessor = useCallback((data: Datum): number => +data[dataKey], [dataKey])
    const formatXLabel = useCallback((index: number): string => (data[index][xDataKey] as unknown) as string, [
        data,
        xDataKey,
    ])

    // Create x (band) d3 scale (see https://observablehq.com/@d3/d3-scaleband)
    // used below to place x axis label and bars in right position on the chart
    const xScale = useMemo(
        () =>
            scaleBand({
                range: [0, innerWidth],
                round: true,
                domain: range(data.length),
                padding: 0.2,
            }),
        [innerWidth, data]
    )

    // Create y linear d3 scale (see https://observablehq.com/@d3/d3-scalelinear)
    // used below to calculate bar height according data and inner height of the chart
    const yScale = useMemo(
        () =>
            scaleLinear({
                range: [innerHeight, 0],
                round: true,
                nice: true,
                domain: [0, Math.max(...data.map(yAccessor))],
            }),
        [innerHeight, data, yAccessor]
    )

    // handlers
    const handleMouseLeave = (): void => {
        tooltipTimeout = window.setTimeout(() => {
            hideTooltip()
        }, 300)
    }

    if (locked) {
        return <LockedChart />
    }

    return (
        <div className={styles.barChart}>
            <svg aria-label="Bar chart" width={width} height={height}>
                <Group left={DEFAULT_PADDING.left} top={DEFAULT_PADDING.top}>
                    <Group aria-label="Chart axes">
                        {/* eslint-disable-next-line jsx-a11y/aria-role */}
                        <Group role="graphics-axis" aria-label="X axis" aria-orientation="horizontal">
                            <AxisBottom
                                top={innerHeight}
                                scale={xScale}
                                tickFormat={formatXLabel}
                                axisLineClassName={styles.axisLine}
                                tickClassName={styles.axisTick}
                            />
                        </Group>

                        {/* eslint-disable-next-line jsx-a11y/aria-role */}
                        <Group role="graphics-axis" aria-orientation="vertical" aria-label="Y axis">
                            <AxisLeft
                                scale={yScale}
                                axisLineClassName={classNames(styles.axisLine, styles.axisLineVertical)}
                                tickClassName={classNames(styles.axisTick, styles.axisTickVertical)}
                            />

                            <GridRows
                                aria-hidden={true}
                                scale={yScale}
                                width={innerWidth}
                                height={innerHeight}
                                className={styles.grid}
                            />
                        </Group>
                    </Group>

                    {/* eslint-disable-next-line jsx-a11y/aria-role */}
                    <Group role="graphics-datagroup" aria-label="Bars">
                        {data.map((datum, index) => {
                            const barHeight = innerHeight - (yScale(yAccessor(datum)) ?? 0)
                            const link = linkURLs?.[index]
                            const classes = classNames(styles.bar, { [styles.barWithLink]: link })
                            const yValue = yAccessor(datum)
                            const xValue = formatXLabel(index)
                            const ariaLabel = `Bar ${index + 1} of ${
                                data.length
                            }. X value: ${xValue}. Y value: ${yValue}`

                            return (
                                <MaybeLink
                                    key={`bar-${index}`}
                                    to={linkURLs?.[index]}
                                    target="_blank"
                                    rel="noopener"
                                    onClick={onDatumLinkClick}
                                    role={linkURLs?.[index] ? 'link' : 'graphics-dataunit'}
                                    aria-label={ariaLabel}
                                    className={styles.barLink}
                                >
                                    <Bar
                                        className={classes}
                                        x={xScale(index)}
                                        y={innerHeight - barHeight}
                                        height={barHeight}
                                        width={xScale.bandwidth()}
                                        fill={fill}
                                        onMouseLeave={handleMouseLeave}
                                        onMouseMove={event => {
                                            if (tooltipTimeout) {
                                                clearTimeout(tooltipTimeout)
                                            }

                                            const rectangle = localPoint(event)

                                            showTooltip({
                                                tooltipData: { xLabel: formatXLabel(index), value: yAccessor(datum) },
                                                tooltipTop: rectangle?.y,
                                                tooltipLeft: rectangle?.x,
                                            })
                                        }}
                                    />
                                </MaybeLink>
                            )
                        })}
                    </Group>
                </Group>
            </svg>

            {tooltipOpen && tooltipData && (
                <TooltipWithBounds className={styles.tooltip} top={tooltipTop} left={tooltipLeft}>
                    <div>
                        <strong>{tooltipData.xLabel}</strong>
                    </div>

                    <div>{tooltipData.value}</div>
                </TooltipWithBounds>
            )}
        </div>
    )
}
