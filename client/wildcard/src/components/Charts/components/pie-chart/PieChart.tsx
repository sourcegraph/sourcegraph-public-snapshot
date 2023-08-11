import { type ReactElement, type SVGProps, useMemo, useState } from 'react'

import { Group } from '@visx/group'
import Pie, { type PieArcDatum } from '@visx/shape/lib/shapes/Pie'
import classNames from 'classnames'
import { noop } from 'rxjs'

import { MaybeLink } from '../../core'
import type { CategoricalLikeChart } from '../../types'

import { PieArc } from './components/PieArc'

import styles from './PieChart.module.scss'

const DEFAULT_PADDING = { top: 20, right: 20, bottom: 20, left: 20 }

/**
 * Generate percent value for each part of pie chart.
 * For example, we have two pies with 30 and 90 values. This means that percentage
 * for the first one is 30 / (30 + 90) = 0.25%
 * and for the second one 90 / (30 + 90) = 0.75%
 */
function getSubtitle<Datum>(arc: PieArcDatum<Datum>, total: number): string {
    return `${((100 * arc.value) / total).toFixed(2)}%`
}

export interface PieChartProps<Datum> extends CategoricalLikeChart<Datum>, SVGProps<SVGSVGElement> {
    width: number
    height: number
    padding?: typeof DEFAULT_PADDING
}

export function PieChart<Datum>(props: PieChartProps<Datum>): ReactElement | null {
    const {
        width,
        height,
        data,
        padding = DEFAULT_PADDING,
        getDatumValue,
        getDatumName,
        getDatumColor,
        getDatumLink = noop,
        onDatumLinkClick = noop,
        className,
        ...attributes
    } = props

    // We have to track which arc is hovered to change order of rendering.
    // Due to the fact that SVG elements don't have CSS z-index (in SVG only
    // order of rendering matters) we have to render additional PieArcs on
    // top of other arcs to visually bring hovered arcs to the top layer.
    const [hoveredArc, setHoveredArc] = useState<PieArcDatum<Datum> | null>(null)
    const sortedData = useMemo(
        () => [...data].sort((first, second) => getDatumValue(second) - getDatumValue(first)),
        [data, getDatumValue]
    )

    const innerWidth = width - padding.left - padding.right
    const innerHeight = height - padding.top - padding.bottom

    const radius = Math.min(innerWidth, innerHeight) / 3
    const centerY = innerHeight / 2
    const centerX = innerWidth / 2

    // Calculate total value (used in PieArc component to get percent value for particular arc)
    const total = useMemo(
        () =>
            sortedData.reduce(
                // Here we have to cast datum[dataKey] to number because we ts can't derive value by dataKey
                (sum, datum) => sum + getDatumValue(datum),
                0
            ),
        [sortedData, getDatumValue]
    )

    if (width < 10) {
        return null
    }

    return (
        <svg
            {...attributes}
            aria-label="Pie chart"
            role="group"
            width={width}
            height={height}
            className={classNames(styles.svg, className)}
        >
            <Group top={centerY + padding.top} left={centerX + padding.left}>
                <Pie data={sortedData} pieValue={getDatumValue} outerRadius={radius} cornerRadius={3}>
                    {pie => (
                        <>
                            <Group role="list">
                                {pie.arcs.map((arc, index) => {
                                    const name = getDatumName(arc.data)
                                    const subtitle = getSubtitle(arc, total)
                                    const link = getDatumLink(arc.data)

                                    return (
                                        <Group key={getDatumName(arc.data)} role="listitem">
                                            <MaybeLink
                                                to={link}
                                                target="_blank"
                                                rel="noopener"
                                                role={link ? 'link' : 'graphics-dataunit'}
                                                aria-label={`Name: ${name}. Value: ${subtitle}.`}
                                                className={classNames(styles.link, { [styles.linkFade]: !!hoveredArc })}
                                                onClick={event => onDatumLinkClick(event, arc.data, index)}
                                                onPointerEnter={() => setHoveredArc(arc)}
                                                onPointerLeave={() => setHoveredArc(null)}
                                                onFocus={() => setHoveredArc(arc)}
                                                onBlur={() => setHoveredArc(null)}
                                            >
                                                <PieArc
                                                    arc={arc}
                                                    path={pie.path}
                                                    title={getDatumName(arc.data)}
                                                    subtitle={subtitle}
                                                    className={styles.arcPath}
                                                    getColor={getDatumColor}
                                                />
                                            </MaybeLink>
                                        </Group>
                                    )
                                })}

                                {hoveredArc && (
                                    <PieArc
                                        aria-hidden={true}
                                        pointerEvents="none"
                                        arc={hoveredArc}
                                        path={pie.path}
                                        title={getDatumName(hoveredArc.data)}
                                        subtitle={getSubtitle(hoveredArc, total)}
                                        className={classNames(styles.arcPath, styles.arcPathFake)}
                                        getColor={getDatumColor}
                                    />
                                )}
                            </Group>
                        </>
                    )}
                </Pie>
            </Group>
        </svg>
    )
}
