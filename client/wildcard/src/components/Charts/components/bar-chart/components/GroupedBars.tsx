import { ComponentProps, MouseEvent, ReactElement, useMemo } from 'react'

import { Group } from '@visx/group'
import { scaleBand } from '@visx/scale'
import { ScaleBand, ScaleLinear } from 'd3-scale'

import { getBrowserName } from '@sourcegraph/common'

import { MaybeLink } from '../../../core'
import { ActiveSegment } from '../types'
import { Category } from '../utils/get-grouped-categories'

import styles from './GroupedBars.module.scss'

interface GroupedBarsProps<Datum> extends ComponentProps<typeof Group> {
    activeSegment: ActiveSegment<Datum> | null
    categories: Category<Datum>[]
    height: number
    xScale: ScaleBand<string>
    yScale: ScaleLinear<number, number>
    getDatumName: (datum: Datum) => string
    getDatumValue: (datum: Datum) => number
    getDatumColor: (datum: Datum) => string | undefined
    getDatumFadeColor?: (datum: Datum) => string
    getDatumLink: (datum: Datum) => string | undefined | null
    onBarHover: (datum: Datum, category: Category<Datum>) => void
    onBarLeave: () => void
    onBarClick: (event: MouseEvent, datum: Datum, index: number) => void
}

const isSafari = getBrowserName() === 'safari'

export function GroupedBars<Datum>(props: GroupedBarsProps<Datum>): ReactElement {
    const {
        width,
        activeSegment,
        categories,
        height,
        xScale,
        yScale,
        getDatumName,
        getDatumValue,
        getDatumColor,
        getDatumFadeColor,
        getDatumLink,
        onBarHover,
        onBarLeave,
        onBarClick,
        ...attributes
    } = props

    const xCategoriesScale = useMemo(
        () =>
            scaleBand<string>({
                domain: [...new Set(categories.flatMap(category => category.data.map(getDatumName)))],
                range: [0, xScale.bandwidth()],
                padding: 0.2,
            }),
        [categories, xScale, getDatumName]
    )

    const handleGroupMouseMove = (event: MouseEvent): void => {
        const [datum, category] = getActiveBar({ event, xScale, xCategoriesScale, categories })

        if (category && datum) {
            if (!activeSegment?.datum) {
                onBarHover(datum, category)
                return
            }

            // Do not call onBarHover every time we mouse move over the same datum
            if (getDatumName(activeSegment.datum) !== getDatumName(datum)) {
                onBarHover(datum, category)
            }
        } else {
            onBarLeave()
        }
    }

    const handleGroupClick = (event: MouseEvent): void => {
        const [datum, , index] = getActiveBar({ event, xScale, xCategoriesScale, categories })

        if (datum && index !== null) {
            onBarClick(event, datum, index)
        }
    }

    return (
        <Group {...attributes} pointerEvents="bounding-rect">
            {categories.map(category => (
                <Group key={category.id} left={xScale(category.id)} height={height}>
                    {category.data.map((datum, index) => {
                        const isOneDatumCategory = category.data.length === 1
                        const barWidth = isOneDatumCategory ? xScale.bandwidth() : xCategoriesScale.bandwidth()
                        const barHeight = height - yScale(getDatumValue(datum))
                        const barX = isOneDatumCategory ? 0 : xCategoriesScale(getDatumName(datum))
                        const barY = yScale(getDatumValue(datum))

                        const barColorProps =
                            activeSegment && activeSegment.category.id !== category.id
                                ? getDatumFadeColor
                                    ? { fill: getDatumFadeColor(datum) }
                                    : // We use css filters to calculate lighten/darken color for non-active bars
                                      // CSS filters don't work in Safari for SVG elements, so we fall back on opacity
                                      { className: styles.barFade, opacity: isSafari ? 0.5 : 1 }
                                : {}

                        return (
                            <MaybeLink
                                key={`bar-group-bar-${category.id}-${getDatumName(datum)}`}
                                to={getDatumLink(datum)}
                                onFocus={() => onBarHover(datum, category)}
                                onClick={event => onBarClick(event, datum, index)}
                            >
                                <rect
                                    x={barX}
                                    y={barY}
                                    width={barWidth}
                                    height={barHeight}
                                    rx={2}
                                    fill={getDatumColor(datum)}
                                    {...barColorProps}
                                />
                            </MaybeLink>
                        )
                    })}
                </Group>
            ))}

            <rect
                width={width}
                height={height}
                fill="transparent"
                opacity={0}
                onMouseMove={handleGroupMouseMove}
                onMouseLeave={onBarLeave}
                onClick={handleGroupClick}
            />
        </Group>
    )
}

interface GetActiveBarInput<Datum> {
    event: MouseEvent
    categories: Category<Datum>[]
    xScale: ScaleBand<string>
    xCategoriesScale: ScaleBand<string>
}

function scaleBandInvert(scale: ScaleBand<string>): (x: number) => number {
    const domains = scale.domain()
    const paddingOuter = scale.paddingOuter()
    const eachBand = scale.step()

    return function (value: number) {
        return Math.max(0, Math.min(domains.length - 1, Math.floor((value - paddingOuter) / eachBand)))
    }
}

type ActiveBarTuple<Datum> = [datum: Datum | null, category: Category<Datum> | null, index: number | null]

function getActiveBar<Datum>(input: GetActiveBarInput<Datum>): ActiveBarTuple<Datum> {
    const { event, xCategoriesScale, categories, xScale } = input

    const targetRectangle = (event.currentTarget as Element).getBoundingClientRect()
    const xCord = event.clientX - targetRectangle.left

    const invertX = scaleBandInvert(xScale)
    const categoryPossibleIndex = invertX(xCord)
    const category = categories[categoryPossibleIndex]

    if (!category) {
        return [null, null, null]
    }

    const isOneDatumCategory = category.data.length === 1

    if (isOneDatumCategory) {
        return [category.data[0], category, categoryPossibleIndex]
    }

    const invertCategories = scaleBandInvert(xCategoriesScale)
    const categoryWindow = xScale(category.id) ?? 0
    const possibleBarIndex = invertCategories(xCord - categoryWindow)

    if (category.data[possibleBarIndex]) {
        return [category.data[possibleBarIndex], category, possibleBarIndex]
    }

    return [null, null, null]
}
