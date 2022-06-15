import React, { CSSProperties, SVGProps } from 'react'

import { LineChart, SeriesLikeChart } from '../../../../../../charts'
import { SeriesBasedChartTypes } from '../../types'
import { LockedChart } from '../locked/LockedChart'

export interface SeriesChartProps<D> extends SeriesLikeChart<D>, Omit<SVGProps<SVGSVGElement>, 'type'> {
    type: SeriesBasedChartTypes
    width: number
    height: number
    zeroYAxisMin?: boolean
    locked?: boolean
    isSeriesSelected?: (id: string) => boolean
    isSeriesHovered?: (id: string) => boolean
}

export function SeriesChart<Datum>(props: SeriesChartProps<Datum>): React.ReactElement {
    const { type, locked, isSeriesSelected = () => true, isSeriesHovered = () => true, ...otherProps } = props

    const getOpacity = (id: string, hasActivePoint: boolean, isActive: boolean): number => {
        // Highlight series with active point
        if (hasActivePoint) {
            if (isActive) {
                return 1
            }

            return 0.5
        }

        if (isSeriesSelected(id)) {
            return 1
        }

        if (isSeriesHovered(id)) {
            return 0.5
        }

        return 1
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
            isSeriesSelected={isSeriesSelected}
            isSeriesHovered={isSeriesHovered}
            getLineGroupStyle={getHoverStyle}
            {...otherProps}
        />
    )
}
