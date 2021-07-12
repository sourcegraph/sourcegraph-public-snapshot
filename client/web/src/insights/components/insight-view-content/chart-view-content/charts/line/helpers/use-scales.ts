import { AxisScaleOutput } from '@visx/axis'
import { DefaultOutput, ScaleConfig, scaleLinear, scaleTime } from '@visx/scale'
import { PickScaleConfigWithoutType } from '@visx/scale/lib/types/ScaleConfig'
import { ScaleLinear, ScaleTime } from 'd3-scale'
import { useMemo } from 'react'

import { Accessors } from '../types'

import { getMinAndMax } from './get-min-max'
import { getRangeWithPadding } from './get-range-with-padding'

const LINE_VERTICAL_PADDING = 0.15

interface UseScalesProps<Datum> {
    /**
     * D3 scales configuration
     * See https://github.com/d3/d3-scale#continuous_domain
     */
    config: {
        x: ScaleConfig<AxisScaleOutput, any, any>
        y: ScaleConfig<AxisScaleOutput, any, any>
    }
    /** Accessors map to get value from datum object */
    accessors: Accessors<Datum, keyof Datum>
    /** Chart width in px */
    width: number
    /** Chart height in px */
    height: number
    /** Dataset with all series (lines) of chart */
    data: Datum[]
}

interface UseScalesOutput {
    /** Extended (with new domain) configuration for XYChart */
    config: {
        x: ScaleConfig<AxisScaleOutput, any, any>
        y: ScaleConfig<AxisScaleOutput, any, any>
    }

    /** Time d3 scale (used to calculate x position for ticks label) */
    xScale: ScaleTime<number, number>

    /** Continuous d3 scale (used to calculate y position for ticks label) */
    yScale: ScaleLinear<number, number>
}

/**
 * Hook to generate d3 scales according to chart configuration.
 */
export function useScales<Datum>(props: UseScalesProps<Datum>): UseScalesOutput {
    const { config, accessors, width, height, data } = props

    // Extend origin config with calculated domain with vertical padding
    const scalesConfig = useMemo(
        () => ({
            ...config,
            y: {
                ...config.y,
                domain: getRangeWithPadding(getMinAndMax(data, accessors), LINE_VERTICAL_PADDING),
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
