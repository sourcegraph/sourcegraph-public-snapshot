import { type ComponentProps, type MouseEvent, type ReactElement, useMemo, useRef } from 'react'

import { Group } from '@visx/group'
import { scaleBand } from '@visx/scale'
import type { ScaleBand, ScaleLinear } from 'd3-scale'

import { getBrowserName } from '@sourcegraph/common'

import { MaybeLink } from '../../../core'
import type { ActiveSegment } from '../types'
import type { Category } from '../utils/get-grouped-categories'

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
    onBarHover: (datum: Datum, category: Category<Datum>, node: Element) => void
    onBarLeave: () => void
    onBarClick: (event: MouseEvent, datum: Datum, index: number) => void
    onBarFocus: (datum: Datum, category: Category<Datum>, node: Element) => void
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
        onBarFocus,
        ...attributes
    } = props

    const rootRef = useRef<SVGGElement>(null)

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
            const datumName = getDatumName(datum)
            const element = rootRef.current?.querySelector<Element>(`[data-id="${getBarId(category.id, datumName)}"]`)

            if (!element) {
                return
            }

            if (!activeSegment?.datum) {
                onBarHover(datum, category, element)
                return
            }

            // Do not call onBarHover every time we mouse move over the same datum
            if (getDatumName(activeSegment.datum) !== datumName) {
                onBarHover(datum, category, element)
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
        <Group
            {...attributes}
            innerRef={rootRef}
            pointerEvents="bounding-rect"
            aria-label="Bar chart content"
            role="list"
        >
            {categories.map(category => (
                <Group key={category.id} left={xScale(category.id)} height={height}>
                    {category.data.map((datum, index) => {
                        const isOneDatumCategory = category.data.length === 1
                        const value = getDatumValue(datum)
                        const name = getDatumName(datum)
                        const barWidth = isOneDatumCategory ? xScale.bandwidth() : xCategoriesScale.bandwidth()
                        const barHeight = height - yScale(value)
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

                        const barLabelText = `${name}, Value: ${value}`
                        const barCategoryLabelText = isOneDatumCategory
                            ? barLabelText
                            : `Category: ${category.id}, ${barLabelText}`

                        return (
                            <MaybeLink
                                key={`${category.id}-${name}`}
                                data-id={getBarId(category.id, name)}
                                to={getDatumLink(datum)}
                                onFocus={event => onBarFocus(datum, category, event.target)}
                                onClick={event => onBarClick(event, datum, index)}
                                aria-label={barCategoryLabelText}
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
                aria-hidden={true}
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

function getBarId(categoryId: string, datumName: string): string {
    return encodeURIComponent(`${categoryId}${datumName}`)
}
