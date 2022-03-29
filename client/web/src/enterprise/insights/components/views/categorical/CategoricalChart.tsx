import React from 'react'

import { CategoricalLikeChart, PieChart } from '../../../../../charts'
import { CategoricalBasedChartTypes } from '../types'

interface CategoricalChartProps<Datum> extends CategoricalLikeChart<Datum> {
    type: CategoricalBasedChartTypes
    width: number
    height: number
}

export function CategoricalChart<Datum>(props: CategoricalChartProps<Datum>): React.ReactElement | null {
    const { type, ...otherProps } = props

    if (type === CategoricalBasedChartTypes.Pie) {
        return <PieChart {...otherProps} />
    }

    // Bar categorical chart will be implemented later
    return null
}
