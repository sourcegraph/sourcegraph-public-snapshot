import * as React from 'react'
import { ReactElement } from 'react'

import { Group } from '@visx/group'
import { scaleLinear } from '@visx/scale'
import { BarRounded } from '@visx/shape'

interface LinearPieChartProps<Datum> {
    width: number
    height: number
    data: Datum[]
    viewMinMax?: [number, number]
    minBarWidth?: number
    getDatumName: (datum: Datum) => string
    getDatumValue: (datum: Datum) => number
    getDatumColor: (datum: Datum) => string | undefined
    className?: string
    rightToLeft?: boolean
    children?: ({ path, datum }: { path: string; datum: Datum }) => React.ReactNode
    barRadius?: number
}

// TODO: rename StackedHorizontalBar
export function LinearPieChart<Datum>({
    className,
    width,
    height,
    data,
    viewMinMax,
    getDatumValue,
    getDatumName,
    getDatumColor,
    children,
    rightToLeft,
    minBarWidth,
    barRadius,
}: LinearPieChartProps<Datum>): ReactElement | null {
    if (width === 0) {
        return null
    }
    const minMax = viewMinMax || [0, data.length > 0 ? data.map(getDatumValue).reduce((sum, next) => sum + next, 0) : 1]
    const xScale = scaleLinear<number>({
        domain: minMax,
        range: [0, width],
    })

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

                    const name = getDatumName(stackedDatum.datum)
                    const color = getDatumColor(stackedDatum.datum)

                    return (
                        <BarRounded
                            key={`bar-group-bar-${name}`}
                            x={barXEnd - barWidth}
                            y={barY}
                            width={barWidth}
                            height={barHeight}
                            fill={color}
                            radius={barRadius || 0}
                            left={isFirstBar}
                            right={isLastBar}
                        >
                            {children && (({ path }) => children({ path, datum: stackedDatum.datum }))}
                        </BarRounded>
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
        const previousStackedValue = stack.length !== 0 ? stack[stack.length - 1].stackedValue : 0

        stack.push({ datum: item, stackedValue: previousStackedValue + getDatumValue(item) })

        return stack
    }, [])
}
