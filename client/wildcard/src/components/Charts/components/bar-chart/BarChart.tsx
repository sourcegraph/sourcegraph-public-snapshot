import { type ReactElement, type SVGProps, useMemo, type MouseEvent } from 'react'

import { scaleBand, scaleLinear } from '@visx/scale'
import type { ScaleBand } from 'd3-scale'

import { SvgAxisBottom, SvgAxisLeft, SvgContent, SvgRoot } from '../../core'
import type { GetScaleTicksOptions } from '../../core/components/axis/tick-formatters'
import type { CategoricalLikeChart } from '../../types'

import { BarChartContent } from './BarChartContent'
import { getGroupedCategories } from './utils/get-grouped-categories'

const DEFAULT_LINK_GETTER = (): null => null

export interface BarChartProps<Datum> extends CategoricalLikeChart<Datum>, SVGProps<SVGSVGElement> {
    width: number
    height: number
    stacked?: boolean
    sortByValue?: boolean

    // TODO: Move these specific only to the axis label UI props to the axis components
    // see https://github.com/sourcegraph/sourcegraph/issues/40009
    pixelsPerYTick?: number
    pixelsPerXTick?: number
    hideXTicks?: boolean
    minAngleXTick?: number
    maxAngleXTick?: number
    getScaleXTicks?: <T>(options: GetScaleTicksOptions) => T[]
    getTruncatedXTick?: (formattedTick: string) => string
    getCategory?: (datum: Datum) => string | undefined
    getDatumFadeColor?: (datum: Datum) => string
    // Provides a lower bound for stretching the Y-axis scale of the chart.
    // By default, when this value is not defined, the chart stretches to the max
    // value of the preseted data. When this value is provided, and higher than
    // any data point, the chart will stretch its scale to this specified value,
    // instead of the highest data point.
    maxValueLowerBound?: number

    onDatumHover?: (datum: Datum) => void
    getDatumHoverValueLabel?: (datum: Datum) => string
}

export function BarChart<Datum>(props: BarChartProps<Datum>): ReactElement {
    const {
        width: outerWidth,
        height: outerHeight,
        data,
        pixelsPerYTick,
        pixelsPerXTick,
        hideXTicks,
        minAngleXTick,
        maxAngleXTick,
        stacked = false,
        sortByValue,
        'aria-label': ariaLabel = 'Bar chart',
        getDatumHover,
        getScaleXTicks,
        getTruncatedXTick,
        getDatumName,
        getDatumValue,
        getDatumColor,
        getDatumFadeColor,
        maxValueLowerBound,
        getDatumHoverValueLabel,
        getDatumLink = DEFAULT_LINK_GETTER,
        getCategory = getDatumName,
        onDatumLinkClick,
        onDatumHover,
        ...attributes
    } = props

    const categories = useMemo(
        () => getGroupedCategories({ data, stacked, sortByValue, getCategory, getDatumName, getDatumValue }),
        [data, stacked, sortByValue, getCategory, getDatumName, getDatumValue]
    )

    const xScale = useMemo(
        () =>
            scaleBand<string>({
                domain: categories.map(category => category.id),
                padding: 0.2,
            }),
        [categories]
    )

    const yScale = useMemo(
        () =>
            scaleLinear<number>({
                domain: [
                    0,
                    categories.reduce(
                        (max, category) => Math.max(max, category.maxValue),
                        maxValueLowerBound ?? -Infinity
                    ),
                ],
            }),
        [categories, maxValueLowerBound]
    )

    const handleBarClick = (event: MouseEvent, datum: Datum, index: number): void => {
        const link = getDatumLink(datum)

        onDatumLinkClick?.(event, datum, index)

        if (!event.isDefaultPrevented() && link) {
            window.open(link)
        }
    }

    return (
        <SvgRoot
            {...attributes}
            width={outerWidth}
            height={outerHeight}
            role="group"
            aria-label={ariaLabel}
            xScale={xScale}
            yScale={yScale}
        >
            <SvgAxisLeft pixelsPerTick={pixelsPerYTick} />
            <SvgAxisBottom
                pixelsPerTick={pixelsPerXTick}
                minRotateAngle={minAngleXTick}
                maxRotateAngle={maxAngleXTick}
                hideTicks={hideXTicks}
                getTruncatedTick={getTruncatedXTick}
                getScaleTicks={getScaleXTicks}
            />

            <SvgContent<ScaleBand<string>, any>>
                {({ yScale, xScale, content }) => (
                    <BarChartContent<Datum>
                        // Visx axis interfaces doesn't support scaleLiner scale in
                        // axisScale interface
                        yScale={yScale}
                        xScale={xScale}
                        width={content.width}
                        height={content.height}
                        top={content.top}
                        left={content.left}
                        stacked={stacked}
                        categories={categories}
                        getDatumHover={getDatumHover}
                        getDatumName={getDatumName}
                        getDatumValue={getDatumValue}
                        getDatumColor={getDatumColor}
                        getDatumFadeColor={getDatumFadeColor}
                        getDatumLink={getDatumLink}
                        onBarClick={handleBarClick}
                        onBarHover={onDatumHover}
                        getDatumHoverValueLabel={getDatumHoverValueLabel}
                    />
                )}
            </SvgContent>
        </SvgRoot>
    )
}
