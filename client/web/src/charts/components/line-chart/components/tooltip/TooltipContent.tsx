import React, { ReactElement, useMemo } from 'react'

import { isDefined } from '@sourcegraph/common'

import { DEFAULT_FALLBACK_COLOR } from '../../../../constants'
import { Series } from '../../../../types'
import { Point as DataPoint } from '../../types'
import { isValidNumber, formatYTick } from '../../utils'

import { getListWindow } from './utils/get-list-window'

import styles from './TooltipContent.module.scss'

export function getLineStroke<Datum>(line: Series<Datum>): string {
    return line?.color ?? DEFAULT_FALLBACK_COLOR
}

const MAX_ITEMS_IN_TOOLTIP = 10

export interface TooltipContentProps<Datum> {
    series: Series<Datum>[]
    activePoint: DataPoint<Datum>
    xAxisKey: keyof Datum
}

/**
 * Display tooltip content for XYChart.
 * It consists of title - datetime for current x point and list of all nearest y points.
 */
export function TooltipContent<Datum>(props: TooltipContentProps<Datum>): ReactElement | null {
    const { activePoint, series, xAxisKey } = props
    const { datum, originalDatum } = activePoint

    const lines = useMemo(() => {
        if (!activePoint) {
            return { window: [], leftRemaining: 0, rightRemaining: 0 }
        }

        const sortedSeries = [...series]
            .map(line => {
                const value = datum[line.dataKey]
                const selfValue = originalDatum[line.dataKey]

                if (!isValidNumber(value) || !isValidNumber(selfValue)) {
                    return
                }

                return { ...line, value, selfValue }
            })
            .filter(isDefined)
            .sort((lineA, lineB) => lineB.value - lineA.value)

        // Find index of hovered point
        const hoveredSeriesIndex = sortedSeries.findIndex(line => line.dataKey === activePoint.seriesKey)

        // Normalize index of hovered point
        const centerIndex = hoveredSeriesIndex !== -1 ? hoveredSeriesIndex : Math.floor(sortedSeries.length / 2)

        return getListWindow(sortedSeries, centerIndex, MAX_ITEMS_IN_TOOLTIP)
    }, [activePoint, series, datum, originalDatum])

    const dateString = new Date(+datum[xAxisKey]).toDateString()

    return (
        <>
            <h3>{dateString}</h3>

            <ul className={styles.tooltipList}>
                {lines.leftRemaining > 0 && <li className={styles.item}>... and {lines.leftRemaining} more</li>}
                {lines.window.map(line => {
                    // In stacked mode each line and datum has its original selfValue
                    // and stacked value which is sum of all data items of lines below
                    const selfValue = formatYTick(line.selfValue)
                    const stackedValue = formatYTick(line.value)
                    const datumKey = activePoint.seriesKey
                    const backgroundColor = datumKey === line.dataKey ? 'var(--secondary-2)' : ''

                    /* eslint-disable react/forbid-dom-props */
                    return (
                        <li key={line.dataKey as string} className={styles.item} style={{ backgroundColor }}>
                            <div style={{ backgroundColor: getLineStroke(line) }} className={styles.mark} />

                            <span className={styles.legendText}>{line?.name ?? 'unknown series'}</span>

                            <span className={styles.legendValue}>
                                {selfValue !== stackedValue ? (
                                    selfValue === null || Number.isNaN(selfValue) ? (
                                        '–'
                                    ) : (
                                        <span className="font-weight-bold">{selfValue}</span>
                                    )
                                ) : null}{' '}
                                {stackedValue === null || Number.isNaN(stackedValue) ? '–' : stackedValue}
                            </span>
                        </li>
                    )
                })}
                {lines.rightRemaining > 0 && <li className={styles.item}>... and {lines.rightRemaining} more</li>}
            </ul>
        </>
    )
}
