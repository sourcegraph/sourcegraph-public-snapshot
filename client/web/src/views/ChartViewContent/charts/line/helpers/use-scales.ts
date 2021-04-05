import { useMemo } from 'react'
import { DefaultOutput, ScaleConfig, scaleLinear, scaleTime } from '@visx/scale'

import { Accessors } from '../types'
import { getRangeWithPadding } from './get-range-with-padding'
import { getMinAndMax } from './get-min-max'
import { AxisScaleOutput } from '@visx/axis'
import { PickScaleConfigWithoutType } from '@visx/scale/lib/types/ScaleConfig'
import { ScaleLinear, ScaleTime } from 'd3-scale'

interface UseScalesProps<Datum> {
    config: {
        x: ScaleConfig<AxisScaleOutput, any, any>
        y: ScaleConfig<AxisScaleOutput, any, any>
    }
    accessors: Accessors<Datum, string>
    width: number
    height: number
    data: Datum[]
}

interface UseScalesOutput {
    config: {
        x: ScaleConfig<AxisScaleOutput, any, any>
        y: ScaleConfig<AxisScaleOutput, any, any>
    }
    xScale: ScaleTime<number, number>
    yScale: ScaleLinear<number, number>
}

export function useScales<Datum>(props: UseScalesProps<Datum>): UseScalesOutput {
    const { config, accessors, width, height, data } = props

    const scalesConfig = useMemo(
        () => ({
            ...config,
            y: {
                ...config.y,
                domain: getRangeWithPadding(getMinAndMax(data, accessors), 0.3),
            },
        }),
        [accessors, data, config]
    )

    const xScale = useMemo(
        () =>
            scaleTime({
                ...scalesConfig.x,
                range: [0, width],
                domain: [accessors.x(data[0]), accessors.x(data[data.length - 1])],
            }),
        [scalesConfig.x, accessors, data, width]
    )

    const yScale = useMemo(
        () =>
            scaleLinear({
                ...(scalesConfig.y as PickScaleConfigWithoutType<'linear', DefaultOutput>),
                range: [height, 0],
            }),
        [scalesConfig.y, height]
    )

    return { config: scalesConfig, xScale, yScale }
}
