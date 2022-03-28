import React from 'react'

import { LineChart } from '../../components/line-chart'
import { Series, SeriesBasedChartTypes } from '../../types'

export interface SeriesChartProps<D> {
    type: SeriesBasedChartTypes
    width: number
    height: number
    data: D[]
    series: Series<D>[]
    xAxisKey: keyof D
    stacked?: boolean
    onDatumClick?: (event: React.MouseEvent) => void
}

export function SeriesChart<Datum>(props: SeriesChartProps<Datum>): React.ReactElement {
    const { type, ...otherProps } = props

    return <LineChart {...otherProps} />
}
