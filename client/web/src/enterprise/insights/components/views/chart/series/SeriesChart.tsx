import React, { CSSProperties, SVGProps, useMemo } from 'react'

import { LineChart, SeriesLikeChart } from '../../../../../../charts'
import { UseSeriesToggleReturn } from '../../../../../../insights/utils/use-series-toggle'
import { SeriesBasedChartTypes } from '../../types'
import { LockedChart } from '../locked/LockedChart'

export interface SeriesChartProps<D> extends SeriesLikeChart<D>, Omit<SVGProps<SVGSVGElement>, 'type'> {
    type: SeriesBasedChartTypes
    width: number
    height: number
    zeroYAxisMin?: boolean
    locked?: boolean
    seriesToggleState?: UseSeriesToggleReturn
}

const FULL_COLOR = 1
const DIMMED_COLOR = 0.5

export function SeriesChart<Datum>(props: SeriesChartProps<Datum>): React.ReactElement {
    const { series, type, locked, seriesToggleState, ...otherProps } = props
    const { isSeriesHovered = () => true, isSeriesSelected = () => true, hoveredId, hasSelections = () => true } =
        seriesToggleState || {}

    const availableSeriesids = useMemo(() => series.map(({ id }) => `${id}`), [series])

    const selectedSeries = useMemo(() => series.filter(({ id }) => isSeriesSelected(`${id}`)), [
        series,
        isSeriesSelected,
    ])

    const getOpacity = (id: string, hasActivePoint: boolean, isActive: boolean): number => {
        if (!hasSelections(availableSeriesids) && hoveredId && !isSeriesHovered(id)) {
            return DIMMED_COLOR
        }

        if (hoveredId && !isSeriesHovered(id)) {
            return DIMMED_COLOR
        }

        // Highlight series with active point
        if (hasActivePoint) {
            if (isActive) {
                return FULL_COLOR
            }

            return DIMMED_COLOR
        }

        if (isSeriesSelected(id)) {
            return FULL_COLOR
        }

        if (isSeriesHovered(id)) {
            return DIMMED_COLOR
        }

        return FULL_COLOR
    }

    const getHoverStyle = (id: string, hasActivePoint: boolean, isActive: boolean): CSSProperties => {
        const opacity = getOpacity(id, hasActivePoint, isActive)

        return {
            opacity,
            transitionProperty: 'opacity',
            transitionDuration: '200ms',
            transitionTimingFunction: 'ease-out',
        }
    }

    if (locked) {
        return <LockedChart />
    }

    return (
        <LineChart
            series={series}
            tooltipSeries={selectedSeries}
            isSeriesSelected={isSeriesSelected}
            isSeriesHovered={isSeriesHovered}
            getLineGroupStyle={getHoverStyle}
            {...otherProps}
        />
    )
}
