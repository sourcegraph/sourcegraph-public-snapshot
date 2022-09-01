import { ReactElement, SVGProps, useMemo, useState } from 'react'

import { Group } from '@visx/group'
import Pie, { PieArcDatum } from '@visx/shape/lib/shapes/Pie'
import classNames from 'classnames'
import { noop } from 'rxjs'

import { MaybeLink } from '../../core'
import { CategoricalLikeChart } from '../../types'

import { PieArc } from './components/PieArc'
import { distributePieArcs } from './distribute-pie-data'

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
    // Due to the fact that svg elements don't have css z-index (in svg only order of renderings matters)
    // we have to render PieArcs in different order to prevent visual label overlapping on small
    // datasets. When user hovers one pie arc we change arcs order in a way to put this
    // hovered arc last in arcs array. By that we sort of change z-index (ordering) of svg element
    // and put hovered label and arc over other arc elements
    const [hoveredArc, setHoveredArc] = useState<PieArcDatum<Datum> | null>(null)
    const sortedData = useMemo(() => distributePieArcs(data, getDatumValue), [data, getDatumValue])

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
            width={width}
            height={height}
            className={classNames(styles.svg, className)}
        >
            <Group top={centerY + padding.top} left={centerX + padding.left}>
                <Pie data={sortedData} pieValue={getDatumValue} outerRadius={radius} cornerRadius={3}>
                    {pie => {
                        const arcs = hoveredArc
                            ? [...pie.arcs.filter(arc => arc.index !== hoveredArc?.index), hoveredArc]
                            : pie.arcs

                        return (
                            <Group>
                                {arcs.map((arc, index) => (
                                    <MaybeLink
                                        key={getDatumName(arc.data)}
                                        to={getDatumLink(arc.data)}
                                        target="_blank"
                                        rel="noopener"
                                        className={styles.link}
                                        role={getDatumLink(arc.data) ? 'link' : 'graphics-dataunit'}
                                        aria-label={`Element ${index + 1} of ${arcs.length}. Name: ${getDatumName(
                                            arc.data
                                        )}. Value: ${getSubtitle(arc, total)}.`}
                                        onClick={event => onDatumLinkClick(event, arc.data)}
                                    >
                                        <PieArc
                                            arc={arc}
                                            path={pie.path}
                                            title={getDatumName(arc.data)}
                                            subtitle={getSubtitle(arc, total)}
                                            className={classNames(styles.arcPath, {
                                                [styles.arcPathWithLink]: !!getDatumLink(arc.data),
                                            })}
                                            getColor={getDatumColor}
                                            onPointerMove={() => setHoveredArc(arc)}
                                        />
                                    </MaybeLink>
                                ))}
                            </Group>
                        )
                    }}
                </Pie>
            </Group>
        </svg>
    )
}
