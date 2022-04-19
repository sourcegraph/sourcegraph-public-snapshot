import React, { SVGProps } from 'react'

import { CategoricalLikeChart, PieChart } from '../../../../../../charts'
import { CategoricalBasedChartTypes } from '../../types'
import { LockedChart } from '../locked/LockedChart'

export interface CategoricalChartProps<Datum>
    extends CategoricalLikeChart<Datum>,
        Omit<SVGProps<SVGSVGElement>, 'type'> {
    type: CategoricalBasedChartTypes
    width: number
    height: number
    locked?: boolean
}

export function CategoricalChart<Datum>(props: CategoricalChartProps<Datum>): React.ReactElement | null {
    const { type, locked, ...otherProps } = props

    if (locked) {
        return <LockedChart />
    }

    if (type === CategoricalBasedChartTypes.Pie) {
        return <PieChart {...otherProps} />
    }

    // Bar categorical chart will be implemented later
    return null
}
