import { ReactElement, useMemo } from 'react'

import { isDefined } from '@sourcegraph/common'
import { H3 } from '@sourcegraph/wildcard'

import { Point } from '../../types'
import { isValidNumber, formatYTick, SeriesWithData, SeriesDatum, getDatumValue } from '../../utils'

import { getListWindow } from './utils/get-list-window'

import styles from './TooltipContent.module.scss'

const MAX_ITEMS_IN_TOOLTIP = 10

export type MinimumPointInfo<Datum> = Pick<Point<Datum>, 'seriesId' | 'value' | 'time'>

export interface TooltipContentProps<Datum> {
    series: SeriesWithData<Datum>[]
    activePoint: MinimumPointInfo<Datum>
    stacked: boolean
}

/**
 * Display tooltip content for XYChart.
 * It consists of title - datetime for current x point and list of all nearest y points.
 */
export function TooltipContent<Datum>(props: TooltipContentProps<Datum>): ReactElement | null {
    const { activePoint, series, stacked } = props

    const lines = useMemo(() => {
        if (!activePoint) {
            return { window: [], leftRemaining: 0, rightRemaining: 0 }
        }

        const sortedSeries = series
            .map(line => {
                const seriesDatum = (line.data as SeriesDatum<Datum>[]).find(
                    datum => datum.x.getTime() === activePoint.time.getTime()
                )
                const value = seriesDatum ? getDatumValue(seriesDatum) : null

                if (!isValidNumber(value)) {
                    return
                }

                return { ...line, value }
            })
            .filter(isDefined)
            .sort((lineA, lineB) => (!stacked ? lineB.value - lineA.value : -1))

        // Find index of hovered point
        const hoveredSeriesIndex = sortedSeries.findIndex(line => line.id === activePoint.seriesId)

        // Normalize index of hovered point
        const centerIndex = hoveredSeriesIndex !== -1 ? hoveredSeriesIndex : Math.floor(sortedSeries.length / 2)

        return getListWindow(sortedSeries, centerIndex, MAX_ITEMS_IN_TOOLTIP)
    }, [activePoint, series, stacked])

    return (
        <>
            <H3>{activePoint.time.toDateString()}</H3>

            <ul className={styles.tooltipList}>
                {lines.leftRemaining > 0 && <li className={styles.item}>... and {lines.leftRemaining} more</li>}
                {lines.window.map(line => {
                    const value = formatYTick(line.value)
                    const isActiveLine = activePoint.seriesId === line.id
                    const stackedValue = isActiveLine && stacked ? formatYTick(activePoint.value) : null
                    const backgroundColor = isActiveLine ? 'var(--secondary-2)' : ''

                    /* eslint-disable react/forbid-dom-props */
                    return (
                        <li key={line.id} className={styles.item} style={{ backgroundColor }}>
                            <div style={{ backgroundColor: getLineColor(line) }} className={styles.mark} />

                            <span className={styles.legendText}>{line.name}</span>

                            {stackedValue && (
                                <span className={styles.legendStackedValue}>
                                    {stackedValue}
                                    {'\u00A0—\u00A0'}
                                </span>
                            )}

                            <span>{value}</span>
                        </li>
                    )
                })}
                {lines.rightRemaining > 0 && <li className={styles.item}>... and {lines.rightRemaining} more</li>}
            </ul>
        </>
    )
}
