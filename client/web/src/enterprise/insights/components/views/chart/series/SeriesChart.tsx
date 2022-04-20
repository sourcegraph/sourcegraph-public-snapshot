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
}

export function SeriesChart<Datum>(props: SeriesChartProps<Datum>): React.ReactElement {
    const { type, locked, ...otherProps } = props

    if (locked) {
        return <LockedChart />
    }

    return <LineChart {...otherProps} />
}
