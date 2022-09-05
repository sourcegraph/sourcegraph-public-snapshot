import { ComponentProps, MouseEvent, ReactElement } from 'react'

import { Group } from '@visx/group'
import { BarRounded } from '@visx/shape'
import { ScaleBand, ScaleLinear } from 'd3-scale'

import { Category } from '../utils/get-grouped-categories'

interface StackedBarsProps<Datum> extends ComponentProps<typeof Group> {
    xScale: ScaleBand<string>
    yScale: ScaleLinear<number, number>
    categories: Category<Datum>[]
    height: number
    getDatumName: (datum: Datum) => string
    getDatumValue: (datum: Datum) => number
    getDatumColor: (datum: Datum) => string | undefined
    onBarHover: (datum: Datum, category: Category<Datum>) => void
    onBarLeave: () => void
    onBarClick: (event: MouseEvent, datum: Datum, index: number) => void
}

export function StackedBars<Datum>(props: StackedBarsProps<Datum>): ReactElement {
    const {
        xScale,
        yScale,
        categories,
        height,
        getDatumValue,
        getDatumName,
        getDatumColor,
        onBarHover,
        onBarClick,
        onBarLeave,
        ...attributes
    } = props

    const stackedCategories = getStackedData({ getDatumValue, categories })

    return (
        <Group {...attributes}>
            {stackedCategories.map(category => (
                <Group key={category.id} left={xScale(category.id)}>
                    {category.stackedData.map((stackedDatum, index) => {
                        const isFirstBar = index === 0
                        const isLastBar = category.stackedData.length - 1 === index
                        const barWidth = xScale.bandwidth()
                        const barX = 0
                        const barHeight = height - yScale(getDatumValue(stackedDatum.datum))
                        const barY = yScale(stackedDatum.stackedValue)

                        return (
                            <BarRounded
                                key={`bar-group-bar-${category.id}-${getDatumName(stackedDatum.datum)}`}
                                x={barX}
                                y={barY}
                                width={barWidth}
                                height={barHeight}
                                fill={getDatumColor(stackedDatum.datum)}
                                radius={5}
                                bottom={isFirstBar}
                                top={isLastBar}
                                onMouseEnter={() => onBarHover(stackedDatum.datum, category)}
                                onClick={event => onBarClick(event, stackedDatum.datum, index)}
                                onMouseLeave={onBarLeave}
                            />
                        )
                    })}
                </Group>
            ))}
        </Group>
    )
}

interface GetStackedDataInput<Datum> {
    categories: Category<Datum>[]
    getDatumValue: (datum: Datum) => number
}

export interface StackedCategory<Datum> extends Category<Datum> {
    stackedData: StackedDatum<Datum>[]
}

interface StackedDatum<Datum> {
    datum: Datum
    stackedValue: number
}

function getStackedData<Datum>(input: GetStackedDataInput<Datum>): StackedCategory<Datum>[] {
    const { categories, getDatumValue } = input

    return categories.map<StackedCategory<Datum>>(category => ({
        ...category,
        stackedData: category.data.reduce<StackedDatum<Datum>[]>((stack, item) => {
            const previousStackedValue = stack.length !== 0 ? stack[stack.length - 1].stackedValue : 0

            stack.push({ datum: item, stackedValue: previousStackedValue + getDatumValue(item) })

            return stack
        }, []),
    }))
}
