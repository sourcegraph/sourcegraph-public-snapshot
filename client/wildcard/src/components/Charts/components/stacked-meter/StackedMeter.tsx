import type { ReactElement } from 'react'

import { Group } from '@visx/group'
import { scaleLinear } from '@visx/scale'
import { BarRounded } from '@visx/shape'

interface StackedMeterProps<Datum> {
    width: number
    height: number
    data: Datum[]
    viewMinMax?: [number, number]
    minBarWidth?: number
    getDatumName: (datum: Datum) => string
    getDatumValue: (datum: Datum) => number
    getDatumColor?: (datum: Datum) => string
    getDatumClassName?: (datum: Datum) => string
    className?: string
    rightToLeft?: boolean
}

export function StackedMeter<Datum>({
    className,
    width,
    height,
    data,
    viewMinMax,
    getDatumValue,
    getDatumName,
    getDatumColor,
    getDatumClassName,
    rightToLeft,
    minBarWidth,
}: StackedMeterProps<Datum>): ReactElement | null {
    if (width === 0) {
        return null
    }
    const minMax = viewMinMax || [0, data.length > 0 ? data.map(getDatumValue).reduce((sum, next) => sum + next, 0) : 1]
    const xScale = scaleLinear<number>({
        domain: minMax,
        range: [0, width],
    })

    // ensure the bars have some minimum width, for aesthetic reasons
    const adjustedGetDatumValue: (datum: Datum) => number = minBarWidth
        ? (datum: Datum): number => Math.max(getDatumValue(datum), ((minMax[1] - minMax[0]) * minBarWidth) / width)
        : getDatumValue

    const stackedData = getStackedData({ getDatumValue: adjustedGetDatumValue, data })

    return (
        <svg width={width} height={height} className={className}>
            <Group top={0} left={0} transform={rightToLeft ? `scale(-1, 1) translate(-${width}, 0)` : undefined}>
                {stackedData.map((stackedDatum, index) => {
                    const isFirstBar = index === 0
                    const isLastBar = data.length - 1 === index
                    const barHeight = height
                    const barY = 0

                    const value = adjustedGetDatumValue(stackedDatum.datum)
                    const barWidth = xScale(value)
                    const barXEnd = xScale(stackedDatum.stackedValue)

                    return (
                        <BarRounded
                            key={`bar-group-bar-${getDatumName(stackedDatum.datum)}`}
                            x={barXEnd - barWidth}
                            y={barY}
                            width={barWidth}
                            height={barHeight}
                            fill={getDatumColor?.(stackedDatum.datum)}
                            className={getDatumClassName?.(stackedDatum.datum)}
                            radius={5}
                            left={isFirstBar}
                            right={isLastBar}
                        />
                    )
                })}
            </Group>
        </svg>
    )
}

interface StackedDatum<Datum> {
    datum: Datum
    stackedValue: number
}

export function getStackedData<Datum>(input: {
    data: Datum[]
    getDatumValue: (datum: Datum) => number
}): StackedDatum<Datum>[] {
    const { data, getDatumValue } = input

    return data.reduce<StackedDatum<Datum>[]>((stack, item) => {
        const previousStackedValue = stack.length !== 0 ? stack.at(-1)!.stackedValue : 0

        stack.push({ datum: item, stackedValue: previousStackedValue + getDatumValue(item) })

        return stack
    }, [])
}
