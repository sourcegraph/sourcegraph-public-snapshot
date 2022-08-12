import { ReactElement, SVGProps, useMemo } from 'react'

import { scaleBand, scaleLinear } from '@visx/scale'
import { ScaleBand } from 'd3-scale'
import { noop } from 'lodash'

import { SvgAxisBottom, SvgAxisLeft, SvgContent, SvgRoot } from '../../core/components/SvgRoot'
import { CategoricalLikeChart } from '../../types'

import { BarChartContent } from './BarChartContent'
import { getGroupedCategories } from './utils/get-grouped-categories'

const DEFAULT_LINK_GETTER = (): null => null

interface BarChartProps<Datum> extends CategoricalLikeChart<Datum>, SVGProps<SVGSVGElement> {
    width: number
    height: number
    stacked?: boolean
    getCategory?: (datum: Datum) => string | undefined
}

export function BarChart<Datum>(props: BarChartProps<Datum>): ReactElement {
    const {
        width: outerWidth,
        height: outerHeight,
        data,
        stacked = false,
        getDatumName,
        getDatumValue,
        getDatumColor,
        getDatumLink = DEFAULT_LINK_GETTER,
        getCategory = getDatumName,
        onDatumLinkClick = noop,
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

    const handleBarClick = (datum: Datum): void => {
        const link = getDatumLink(datum)

        if (link) {
            window.open(link)
        }

        onDatumLinkClick(datum)
    }

    return (
        <SvgRoot {...attributes} width={outerWidth} height={outerHeight} xScale={xScale} yScale={yScale}>
            <SvgAxisLeft />
            <SvgAxisBottom />

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
