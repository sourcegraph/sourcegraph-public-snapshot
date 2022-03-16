import { useMemo, useContext } from 'react'

import { AxisScale, AxisScaleOutput } from '@visx/axis'
import { DefaultOutput, ScaleConfig, scaleLinear, scaleTime } from '@visx/scale'
import { PickScaleConfigWithoutType } from '@visx/scale/lib/types/ScaleConfig'
import { ScaleTime } from 'd3-scale'

import { LineChartSettingsContext } from '../line-chart-settings-provider'
import { Accessors } from '../types'

import { getMinAndMax } from './get-min-max'

interface ScalesConfiguration {
    x: ScaleConfig<AxisScaleOutput, any, any>
    y: ScaleConfig<AxisScaleOutput, any, any>
}

export interface UseScalesConfiguration<Datum> {
    /**
     * D3 scales configuration
     * See https://github.com/d3/d3-scale#continuous_domain
     */
    config: ScalesConfiguration

    /**
     * Dataset with all series (lines) of chart
     */
    data: Datum[]

    /**
     * Accessors map to get (x, y) value from datum objects
     */
    accessors: Accessors<Datum, keyof Datum>
}

export function useScalesConfiguration<Datum>(props: UseScalesConfiguration<Datum>): ScalesConfiguration {
    const { config, data, accessors } = props
    const { zeroYAxisMin } = useContext(LineChartSettingsContext)

    // Extend origin config with calculated domain with vertical padding
    return useMemo(() => {
        let [min, max] = getMinAndMax(data, accessors)

        if (zeroYAxisMin) {
            min = 0
        }

        // Generate pseudo domain if all values of dataset are equal
        ;[min, max] = min === max ? [max - max / 2, max + max / 2] : [min, max]

        return {
            ...config,
            y: {
                ...config.y,
                domain: [min, max],
            },
        }
    }, [data, accessors, zeroYAxisMin, config])
}

interface UseYScalesProps {
    /**
     * D3 scales configuration
     * See https://github.com/d3/d3-scale#continuous_domain
     */
    config: ScaleConfig<AxisScaleOutput, any, any>
    height: number
}

export function useYScale(props: UseYScalesProps): AxisScale {
    const { config, height } = props

    return useMemo(
        () =>
            scaleLinear({
                ...(config as PickScaleConfigWithoutType<'linear', DefaultOutput>),
                range: [height, 0],
            }),
        [config, height]
    )
}

interface UseXScalesProps<Datum> {
    /**
     * D3 scales configuration
     * See https://github.com/d3/d3-scale#continuous_domain
     */
    config: ScaleConfig<AxisScaleOutput, any, any>

    /** Accessors map to get value from datum object */
    accessors: Accessors<Datum, keyof Datum>

    /** Chart width in px */
    width: number

    /** Dataset with all series (lines) of chart */
    data: Datum[]
}

export function useXScale<Datum>(props: UseXScalesProps<Datum>): ScaleTime<number, number> {
    const { config, accessors, data, width } = props

    return useMemo(
        () =>
            scaleTime({
                ...config,
                range: [0, width],
                domain: [accessors.x(data[0]), accessors.x(data[data.length - 1])],
            }),
        [config, accessors, data, width]
    )
}
