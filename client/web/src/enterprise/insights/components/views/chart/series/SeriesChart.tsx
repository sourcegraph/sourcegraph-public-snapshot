import React, { SVGProps } from 'react'

import { LineChart, SeriesLikeChart } from '../../../../../../charts'
import { SeriesBasedChartTypes } from '../../types'

export interface SeriesChartProps<D> extends SeriesLikeChart<D>, Omit<SVGProps<SVGSVGElement>, 'type'> {
    type: SeriesBasedChartTypes
    width: number
    height: number
}

export function SeriesChart<Datum>(props: SeriesChartProps<Datum>): React.ReactElement {
    const { type, ...otherProps } = props

    return <LineChart {...otherProps} />
}
