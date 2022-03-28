import React from 'react'

import { PieChart } from '../../components/pie-chart'
import { CategoricalBasedChartTypes } from '../../types'

interface CategoricalChartProps<Datum> {
    type: CategoricalBasedChartTypes
    width: number
    height: number
    data: Datum[]

    getDatumValue: (datum: Datum) => number
    getDatumName: (datum: Datum) => string
    getDatumColor: (datum: Datum) => string | undefined
    getDatumLink: (datum: Datum) => string | undefined
    onDatumLinkClick?: (event: React.MouseEvent) => void
}

export function CategoricalChart<Datum>(props: CategoricalChartProps<Datum>): React.ReactElement | null {
    const { type, ...otherProps } = props

    if (type === CategoricalBasedChartTypes.Pie) {
        return <PieChart {...otherProps} />
    }

    // Bar categorical chart will be implemented later
    return null
}
