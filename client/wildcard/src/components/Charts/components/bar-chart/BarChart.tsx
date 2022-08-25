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

    // TODO: Move these specific only to the axis label UI props to the axis components
    // see https://github.com/sourcegraph/sourcegraph/issues/40009
    pixelsPerYTick?: number
    pixelsPerXTick?: number
    maxAngleXTick?: number
    getScaleXTicks?: <T>(options: GetScaleTicksOptions) => T[]
    getTruncatedXTick?: (formattedTick: string) => string
    getCategory?: (datum: Datum) => string | undefined
}

export function BarChart<Datum>(props: BarChartProps<Datum>): ReactElement {
    const {
        width: outerWidth,
        height: outerHeight,
        data,
        pixelsPerYTick,
        pixelsPerXTick,
        maxAngleXTick,
        stacked = false,
        getDatumHover,
        getScaleXTicks,
        getTruncatedXTick,
        getDatumName,
        getDatumValue,
        getDatumColor,
        getDatumLink = DEFAULT_LINK_GETTER,
        getCategory = getDatumName,
        onDatumLinkClick,
        ...attributes
    } = props

    const categories = useMemo(
        () => getGroupedCategories({ data, stacked, getCategory, getDatumName, getDatumValue }),
        [data, stacked, getCategory, getDatumName, getDatumValue]
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

    const handleBarClick = (event: MouseEvent, datum: Datum): void => {
        const link = getDatumLink(datum)

        onDatumLinkClick?.(event, datum)

        if (!event.isDefaultPrevented() && link) {
            window.open(link)
        }
    }

    return (
        <SvgRoot {...attributes} width={outerWidth} height={outerHeight} xScale={xScale} yScale={yScale}>
            <SvgAxisLeft pixelsPerTick={pixelsPerYTick} />
            <SvgAxisBottom
                pixelsPerTick={pixelsPerXTick}
                maxRotateAngle={maxAngleXTick}
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
                        getDatumLink={getDatumLink}
                        onBarClick={handleBarClick}
                    />
                )}
            </SvgContent>
        </SvgRoot>
    )
}
