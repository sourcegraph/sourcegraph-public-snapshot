import { ReactElement, SVGProps, useMemo, useState } from 'react'

import { scaleBand, scaleLinear } from '@visx/scale'
import classNames from 'classnames'
import { noop } from 'lodash'

import { AxisBottom, AxisLeft, getChartContentSizes, Tooltip } from '../../core'
import { CategoricalLikeChart } from '../../types'

import { GroupedBars } from './components/GroupedBars'
import { StackedBars } from './components/StackedBars'
import { BarTooltipContent } from './components/TooltipContent'
import { Category, getGroupedCategories } from './utils/get-grouped-categories'

import styles from './BarChart.module.scss'

const DEFAULT_LINK_GETTER = (): null => null

interface BarChartProps<Datum> extends CategoricalLikeChart<Datum>, SVGProps<SVGSVGElement> {
    width: number
    height: number
    stacked?: boolean
    getCategory?: (datum: Datum) => string | undefined
}

interface ActiveSegment<Datum> {
    category: Category<Datum>
    datum: Datum
}

export function BarChart<Datum>(props: BarChartProps<Datum>): ReactElement {
    const {
        width: outerWidth,
        height: outerHeight,
        data,
        stacked = false,
        className,
        getDatumName,
        getDatumValue,
        getDatumColor,
        getDatumLink = DEFAULT_LINK_GETTER,
        getCategory = getDatumName,
        onDatumLinkClick = noop,
        ...attributes
    } = props

    const [yAxisElement, setYAxisElement] = useState<SVGGElement | null>(null)
    const [xAxisReference, setXAxisElement] = useState<SVGGElement | null>(null)
    const [activeSegment, setActiveSegment] = useState<ActiveSegment<Datum> | null>(null)

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

    const categories = useMemo(
        () => getGroupedCategories({ data, stacked, getCategory, getDatumName, getDatumValue }),
        [data, stacked, getCategory, getDatumName, getDatumValue]
    )

    const xScale = useMemo(
        () =>
            scaleBand<string>({
                domain: categories.map(category => category.id),
                range: [0, content.width],
                padding: 0.2,
            }),
        [content, categories]
    )

    const yScale = useMemo(
        () =>
            scaleLinear<number>({
                domain: [0, Math.max(...categories.map(category => category.maxValue))],
                range: [content.height, 0],
            }),
        [content, categories]
    )

    const handleBarClick = (datum: Datum): void => {
        const link = getDatumLink(datum)

        if (link) {
            window.open(link)
        }

        onDatumLinkClick(datum)
    }

    const withActiveLink = activeSegment?.datum ? getDatumLink(activeSegment?.datum) : null

    return (
        <svg
            width={outerWidth}
            height={outerHeight}
            {...attributes}
            className={classNames(className, styles.root, { [styles.rootWithHoveredLinkPoint]: withActiveLink })}
        >
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

            {stacked ? (
                <StackedBars
                    categories={categories}
                    xScale={xScale}
                    yScale={yScale}
                    getDatumName={getDatumName}
                    getDatumValue={getDatumValue}
                    getDatumColor={getDatumColor}
                    left={content.left}
                    top={content.top}
                    height={content.height}
                    onBarHover={(datum, category) => setActiveSegment({ datum, category })}
                    onBarLeave={() => setActiveSegment(null)}
                    onBarClick={handleBarClick}
                />
            ) : (
                <GroupedBars
                    categories={categories}
                    xScale={xScale}
                    yScale={yScale}
                    getDatumName={getDatumName}
                    getDatumValue={getDatumValue}
                    getDatumColor={getDatumColor}
                    getDatumLink={getDatumLink}
                    left={content.left}
                    top={content.top}
                    height={content.height}
                    width={content.width}
                    onBarHover={(datum, category) => setActiveSegment({ datum, category })}
                    onBarLeave={() => setActiveSegment(null)}
                    onBarClick={handleBarClick}
                />
            )}

            {activeSegment && (
                <Tooltip>
                    <BarTooltipContent
                        category={activeSegment.category}
                        activeBar={activeSegment.datum}
                        getDatumColor={getDatumColor}
                        getDatumValue={getDatumValue}
                        getDatumName={getDatumName}
                    />
                </Tooltip>
            )}
        </svg>
    )
}
