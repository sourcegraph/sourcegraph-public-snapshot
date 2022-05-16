import { ReactElement, useMemo } from 'react'

import { RenderTooltipParams } from '@visx/xychart/lib/components/Tooltip'

import { isDefined } from '@sourcegraph/common'
import { Typography } from '@sourcegraph/wildcard'

import { LineChartSeriesWithData, Point } from '../../types'
import { getLineStroke } from '../LineChartContent'

import { getListWindow, ListWindow } from './get-list-window'

import styles from './TooltipContent.module.scss'

const MAX_ITEMS_IN_TOOLTIP = 10

export interface TooltipContentProps<Datum extends object> extends RenderTooltipParams<Point> {
    /** Dataset of series (lines) on the chart. */
    series: LineChartSeriesWithData<Datum>[]
}

/**
 * Display tooltip content for XYChart.
 * It consists of title - datetime for current x point and list of all nearest y points.
 */
export function TooltipContent<Datum extends object>(props: TooltipContentProps<Datum>): ReactElement | null {
    const { tooltipData, series } = props
    const datum = tooltipData?.nearestDatum?.datum
    const nearestSeriesKey = tooltipData?.nearestDatum?.key

    const lines = useMemo<ListWindow<LineChartSeriesWithData<Datum> & { point: Point }>>(() => {
        if (!datum) {
            return { window: [], leftRemaining: 0, rightRemaining: 0 }
        }

        const sortedSeries = [...series]
            .map(line => {
                const point = line.data.find(point => +point.x === +datum.x)

                if (!point) {
                    return
                }

                return { ...line, point }
            })
            .filter(isDefined)
            .sort((lineA, lineB) => (lineB.point.y ?? 0) - (lineA.point?.y ?? 0))

        // Find index of hovered point
        const hoveredSeriesIndex = sortedSeries.findIndex(line => line.dataKey === nearestSeriesKey)

        // Normalize index of hovered point
        const centerIndex = hoveredSeriesIndex !== -1 ? hoveredSeriesIndex : Math.floor(sortedSeries.length / 2)

        return getListWindow(sortedSeries, centerIndex, MAX_ITEMS_IN_TOOLTIP)
    }, [series, datum, nearestSeriesKey])

    if (!datum) {
        return null
    }

    const dateString = new Date(datum.x).toDateString()

    return (
        <>
            <Typography.H3>{dateString}</Typography.H3>

            <ul className={styles.tooltipList}>
                {lines.leftRemaining > 0 && <li className={styles.item}>... and {lines.leftRemaining} more</li>}
                {lines.window.map(line => {
                    const value = line.point.y
                    const datumKey = tooltipData?.nearestDatum?.key

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
