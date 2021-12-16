import { RenderTooltipParams } from '@visx/xychart/lib/components/Tooltip'
import React, { ReactElement, useMemo } from 'react'

import { isDefined } from '@sourcegraph/shared/src/util/types'

import { LineChartSeriesWithData, Point } from '../types'

import { getLineStroke } from './LineChartContent'
import styles from './TooltipContent.module.scss'

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

    const lines = useMemo(() => {
        if (!datum) {
            return []
        }

        return [...series]
            .map(line => {
                const point = line.data.find(point => +point.x === +datum.x)

                if (!point) {
                    return
                }

                return { ...line, point }
            })
            .filter(isDefined)
            .sort((lineA, lineB) => (lineB.point.y ?? 0) - (lineA.point?.y ?? 0))
    }, [series, datum])

    if (!datum) {
        return null
    }

    const dateString = new Date(datum.x).toDateString()

    return (
        <>
            <h3>{dateString}</h3>

            <ul className={styles.tooltipList}>
                {lines.map(line => {
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
            </ul>
        </>
    )
}
