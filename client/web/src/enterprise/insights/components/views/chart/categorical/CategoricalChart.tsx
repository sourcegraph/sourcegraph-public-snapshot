import React, { type SVGProps } from 'react'

import { type CategoricalLikeChart, PieChart } from '@sourcegraph/wildcard'

import { LockedChart } from '../locked/LockedChart'

export enum CategoricalBasedChartTypes {
    Pie,
    Bar,
}

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
