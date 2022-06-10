import { ReactElement, SVGProps, useMemo, useState } from 'react'

import { Group } from '@visx/group'
import { scaleBand, scaleLinear } from '@visx/scale'

import { AxisBottom, AxisLeft, getChartContentSizes } from '../../core'
import { CategoricalLikeChart } from '../../types'

import { getGroupedCategories } from './utils/get-grouped-categories'

interface BarChartProps<Datum> extends CategoricalLikeChart<Datum>, SVGProps<SVGSVGElement> {
    width: number
    height: number
    getCategory?: (datum: Datum) => string | undefined
}

export function BarChart<Datum>(props: BarChartProps<Datum>): ReactElement {
    const {
        width: outerWidth,
        height: outerHeight,
        data,
        getDatumName,
        getDatumValue,
        getDatumColor,
        getDatumLink,
        getCategory = getDatumName,
        ...attributes
    } = props

    const [yAxisElement, setYAxisElement] = useState<SVGGElement | null>(null)
    const [xAxisReference, setXAxisElement] = useState<SVGGElement | null>(null)

    const content = useMemo(
        () =>
            getChartContentSizes({
                width: outerWidth,
                height: outerHeight,
                margin: {
                    top: 16,
                    right: 16,
                    left: yAxisElement?.getBBox().width,
                    bottom: xAxisReference?.getBBox().height,
                },
            }),
        [yAxisElement, xAxisReference, outerWidth, outerHeight]
    )

    const categories = useMemo(() => getGroupedCategories({ data, getCategory, getDatumName }), [
        data,
        getCategory,
        getDatumName,
    ])

    const xScale = useMemo(
        () =>
            scaleBand<string>({
                domain: categories.map(category => category.id),
                range: [0, content.width],
                padding: 0.2,
            }),
        [content, categories]
    )

    const xCategoriesScale = useMemo(
        () =>
            scaleBand<string>({
                domain: [...new Set(categories.flatMap(category => category.data.map(getDatumName)))],
                range: [0, xScale.bandwidth()],
                padding: 0.2,
            }),
        [categories, xScale, getDatumName]
    )

    const yScale = useMemo(
        () =>
            scaleLinear<number>({
                domain: [0, Math.max(...categories.map(category => Math.max(...category.data.map(getDatumValue))))],
                range: [content.height, 0],
            }),
        [content, categories, getDatumValue]
    )

    return (
        <svg width={outerWidth} height={outerHeight} {...attributes}>
            <AxisLeft
                ref={setYAxisElement}
                scale={yScale}
                width={content.width}
                height={content.height}
                top={content.top}
                left={content.left}
            />

            <AxisBottom
                ref={setXAxisElement}
                scale={xScale}
                width={content.width}
                top={content.bottom}
                left={content.left}
            />

            <Group left={content.left} top={content.top}>
                {categories.map(category => (
                    <Group key={category.id} left={xScale(category.id)}>
                        {category.data.map(datum => {
                            const isOneDatumCategory = category.data.length === 1
                            const barWidth = isOneDatumCategory ? xScale.bandwidth() : xCategoriesScale.bandwidth()
                            const barHeight = content.height - yScale(getDatumValue(datum))
                            const barX = isOneDatumCategory ? 0 : xCategoriesScale(getDatumName(datum))
                            const barY = yScale(getDatumValue(datum))

                            return (
                                <rect
                                    key={`bar-group-bar-${category.id}-${getDatumName(datum)}`}
                                    x={barX}
                                    y={barY}
                                    width={barWidth}
                                    height={barHeight}
                                    fill={getDatumColor(datum)}
                                    rx={4}
                                />
                            )
                        })}
                    </Group>
                ))}
            </Group>
        </svg>
    )
}
