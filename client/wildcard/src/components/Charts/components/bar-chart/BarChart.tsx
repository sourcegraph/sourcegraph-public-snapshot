import { ReactElement, SVGProps, useMemo, MouseEvent } from 'react'

import { scaleBand, scaleLinear } from '@visx/scale'
import { ScaleBand } from 'd3-scale'

import { GetScaleTicksOptions } from '../../core/components/axis/tick-formatters'
import { SvgAxisBottom, SvgAxisLeft, SvgContent, SvgRoot } from '../../core/components/SvgRoot'
import { CategoricalLikeChart } from '../../types'

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

    onDatumHover?: (datum: Datum) => void
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
        getDatumHover,
        getScaleXTicks,
        getTruncatedXTick,
        getDatumName,
        getDatumValue,
        getDatumColor,
        getDatumFadeColor,
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
                domain: [0, Math.max(...categories.map(category => category.maxValue))],
            }),
        [categories]
    )

    const handleBarClick = (event: MouseEvent, datum: Datum, index: number): void => {
        const link = getDatumLink(datum)

        onDatumLinkClick?.(event, datum, index)

        if (!event.isDefaultPrevented() && link) {
            window.open(link)
        }
    }

    return (
        <SvgRoot {...attributes} width={outerWidth} height={outerHeight} xScale={xScale} yScale={yScale}>
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
                    />
                )}
            </SvgContent>
        </SvgRoot>
    )
}
