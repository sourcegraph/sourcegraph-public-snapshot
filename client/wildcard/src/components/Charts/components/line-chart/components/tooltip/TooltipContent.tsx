import { type ReactElement, useMemo } from 'react'

import { isDefined } from '../../../../../../utils'
import { H3 } from '../../../../../Typography'
import { TooltipList, TooltipListBlankItem, TooltipListItem } from '../../../../core'
import { formatYTick } from '../../../../core/components/axis/tick-formatters'
import type { Point } from '../../types'
import { isValidNumber, type SeriesWithData, type SeriesDatum, getDatumValue, getLineColor } from '../../utils'

import { getListWindow } from './utils/get-list-window'

const MAX_ITEMS_IN_TOOLTIP = 10

export type MinimumPointInfo = Pick<Point, 'seriesId' | 'yValue' | 'xValue'>

export interface TooltipContentProps<Datum> {
    series: SeriesWithData<Datum>[]
    activePoint: MinimumPointInfo
    stacked: boolean
}

/**
 * Display tooltip content for series-like chart.
 * This tooltip renders title - datetime for hovered/focused x point
 * and its list of all nearest y points.
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
                    datum => datum.x.getTime() === activePoint.xValue.getTime()
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
            <H3>{activePoint.xValue.toDateString()}</H3>

            <TooltipList>
                {lines.leftRemaining > 0 && (
                    <TooltipListBlankItem>... and {lines.leftRemaining} more</TooltipListBlankItem>
                )}
                {lines.window.map(line => {
                    // TODO: Support stacked formatted value
                    const value = formatYTick(line.value)
                    const isActiveLine = activePoint.seriesId === line.id

                    return (
                        <TooltipListItem
                            key={line.id}
                            isActive={isActiveLine}
                            name={line.name}
                            value={value}
                            color={getLineColor(line)}
                        />
                    )
                })}
                {lines.rightRemaining > 0 && (
                    <TooltipListBlankItem>... and {lines.rightRemaining} more</TooltipListBlankItem>
                )}
            </TooltipList>
        </>
    )
}
