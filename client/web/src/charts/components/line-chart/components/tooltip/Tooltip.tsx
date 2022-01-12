import React, { ReactElement, useEffect, useMemo, useState } from 'react'

import { isDefined } from '@sourcegraph/common'

import { LineChartSeries, Point } from '../../types'
import { isValidNumber } from '../../utils/data-guards'
import { formatYTick } from '../../utils/ticks'
import { FloatingPanel, Target } from '../floating-panel/FloatingPanel'

import styles from './Tooltip.module.scss'
import { getListWindow } from './utils/get-list-window'

/**
 * Default value for line color in case if we didn't get color for line from content config.
 */
export const DEFAULT_LINE_STROKE = 'var(--gray-07)'

export const getLineStroke = <Datum extends object>(line: LineChartSeries<Datum>): string =>
    line?.color ?? DEFAULT_LINE_STROKE

interface TooltipProps {
    reference?: Target
}

export const Tooltip: React.FunctionComponent<TooltipProps> = props => {
    const { reference } = props
    const [virtualElement, setVirtualElement] = useState<Target>()

    useEffect(() => {
        function handleMove(event: PointerEvent): void {
            setVirtualElement({
                getBoundingClientRect: () => ({
                    width: 0,
                    height: 0,
                    x: event.clientX,
                    y: event.clientY,
                    top: event.clientY,
                    left: event.clientX,
                    right: event.clientX,
                    bottom: event.clientY,
                }),
            })
        }

        window.addEventListener('pointermove', handleMove)

        return () => {
            window.removeEventListener('pointermove', handleMove)
        }
    }, [])

    useEffect(() => {
        if (!reference) {
            return
        }

        setVirtualElement(reference)
    }, [reference])

    if (!virtualElement) {
        return null
    }

    return (
        <FloatingPanel className={styles.tooltip} target={virtualElement} strategy="fixed" placement="right-start">
            {props.children}
        </FloatingPanel>
    )
}

const MAX_ITEMS_IN_TOOLTIP = 10

export interface TooltipContentProps<Datum extends object> {
    series: LineChartSeries<Datum>[]
    data: Datum[]
    activePoint: Point
    xAxisKey: keyof Datum
}

/**
 * Display tooltip content for XYChart.
 * It consists of title - datetime for current x point and list of all nearest y points.
 */
export function TooltipContent<Datum extends object>(props: TooltipContentProps<Datum>): ReactElement | null {
    const { data, activePoint, series, xAxisKey } = props

    const lines = useMemo(() => {
        if (!activePoint) {
            return { window: [], leftRemaining: 0, rightRemaining: 0 }
        }

        const currentDatum = data[activePoint.index]

        const sortedSeries = [...series]
            .map(line => {
                const value = currentDatum[line.dataKey]

                if (!isValidNumber(value)) {
                    return
                }

                return { ...line, value }
            })
            .filter(isDefined)
            .sort((lineA, lineB) => lineB.value - lineA.value)

        // Find index of hovered point
        const hoveredSeriesIndex = sortedSeries.findIndex(line => line.dataKey === activePoint.seriesKey)

        // Normalize index of hovered point
        const centerIndex = hoveredSeriesIndex !== -1 ? hoveredSeriesIndex : Math.floor(sortedSeries.length / 2)

        return getListWindow(sortedSeries, centerIndex, MAX_ITEMS_IN_TOOLTIP)
    }, [series, activePoint, data])

    const dateString = new Date(+data[activePoint.index][xAxisKey]).toDateString()

    return (
        <>
            <h3>{dateString}</h3>

            <ul className={styles.tooltipList}>
                {lines.leftRemaining > 0 && <li className={styles.item}>... and {lines.leftRemaining} more</li>}
                {lines.window.map(line => {
                    const value = formatYTick(line.value)
                    const datumKey = activePoint.seriesKey
                    const backgroundColor = datumKey === line.dataKey ? 'var(--secondary-2)' : ''

                    /* eslint-disable react/forbid-dom-props */
                    return (
                        <li key={line.dataKey as string} className={styles.item} style={{ backgroundColor }}>
                            <div style={{ backgroundColor: getLineStroke(line) }} className={styles.mark} />

                            <span className={styles.legendText}>{line?.name ?? 'unknown series'}</span>

                            <span className={styles.legendValue}>
                                {' '}
                                {value === null || Number.isNaN(value) ? '–' : value}{' '}
                            </span>
                        </li>
                    )
                })}
                {lines.rightRemaining > 0 && <li className={styles.item}>... and {lines.rightRemaining} more</li>}
            </ul>
        </>
    )
}
