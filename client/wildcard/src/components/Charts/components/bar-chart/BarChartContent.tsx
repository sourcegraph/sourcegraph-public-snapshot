import { MouseEvent, ReactElement, SVGProps, useRef, useState } from 'react'

import { Group } from '@visx/group'
import classNames from 'classnames'
import { ScaleBand, ScaleLinear } from 'd3-scale'

import { Tooltip } from '../../core'

import { GroupedBars } from './components/GroupedBars'
import { StackedBars } from './components/StackedBars'
import { BarTooltipContent } from './components/TooltipContent'
import { ActiveSegment } from './types'
import { Category } from './utils/get-grouped-categories'

import styles from './BarChartContent.module.scss'

interface BarChartContentProps<Datum> extends SVGProps<SVGGElement> {
    stacked: boolean

    top: number
    left: number

    xScale: ScaleBand<string>
    yScale: ScaleLinear<number, number>
    categories: Category<Datum>[]

    getDatumName: (datum: Datum) => string
    getDatumValue: (datum: Datum) => number
    getDatumHover?: (datum: Datum) => string
    getDatumColor: (datum: Datum) => string | undefined
    getDatumFadeColor?: (datum: Datum) => string
    getDatumLink: (datum: Datum) => string | undefined | null
    onBarClick: (event: MouseEvent, datum: Datum, index: number) => void
    onBarHover?: (datum: Datum) => void
}

export function BarChartContent<Datum>(props: BarChartContentProps<Datum>): ReactElement {
    const {
        xScale,
        yScale,
        categories,
        stacked,
        top,
        left,
        width = 0,
        height = 0,
        getDatumHover,
        getDatumName,
        getDatumValue,
        getDatumColor,
        getDatumFadeColor,
        getDatumLink,
        onBarClick,
        onBarHover,
        ...attributes
    } = props

    const rootRef = useRef<SVGGElement>(null)
    const [activeSegment, setActiveSegment] = useState<ActiveSegment<Datum> | null>(null)

    const withActiveLink = activeSegment?.datum ? getDatumLink(activeSegment?.datum) : null

    const handleBarHover = (datum: Datum, category: Category<Datum>): void => {
        setActiveSegment({ datum, category })
        onBarHover?.(datum)
    }

    return (
        <Group
            {...attributes}
            innerRef={rootRef}
            className={classNames(styles.root, { [styles.rootWithHoveredLinkPoint]: withActiveLink })}
        >
            {stacked ? (
                <StackedBars
                    aria-label="chart content group"
                    categories={categories}
                    xScale={xScale}
                    yScale={yScale}
                    getDatumName={getDatumName}
                    getDatumValue={getDatumValue}
                    getDatumColor={getDatumColor}
                    height={+height}
                    onBarHover={handleBarHover}
                    onBarLeave={() => setActiveSegment(null)}
                    onBarClick={onBarClick}
                />
            ) : (
                <GroupedBars
                    aria-label="chart content group"
                    activeSegment={activeSegment}
                    categories={categories}
                    xScale={xScale}
                    yScale={yScale}
                    getDatumName={getDatumName}
                    getDatumValue={getDatumValue}
                    getDatumColor={getDatumColor}
                    getDatumFadeColor={getDatumFadeColor}
                    getDatumLink={getDatumLink}
                    height={+height}
                    width={+width}
                    onBarHover={handleBarHover}
                    onBarLeave={() => setActiveSegment(null)}
                    onBarClick={onBarClick}
                />
            )}

            {activeSegment && rootRef.current && (
                <Tooltip containerElement={rootRef.current}>
                    <BarTooltipContent
                        category={activeSegment.category}
                        activeBar={activeSegment.datum}
                        getDatumColor={getDatumColor}
                        getDatumValue={getDatumValue}
                        getDatumName={getDatumName}
                        getDatumHover={getDatumHover}
                    />
                </Tooltip>
            )}
        </Group>
    )
}
