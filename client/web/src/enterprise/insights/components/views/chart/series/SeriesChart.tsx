import React, { SVGProps } from 'react'

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

    if (locked) {
        return <LockedChart />
    }

    return <LineChart isSeriesSelected={isSeriesSelected} isSeriesHovered={isSeriesHovered} {...otherProps} />
}
