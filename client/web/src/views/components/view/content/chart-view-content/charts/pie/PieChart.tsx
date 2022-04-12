import React, { ReactElement, useMemo, useState } from 'react'

import { Group } from '@visx/group'
import Pie, { PieArcDatum } from '@visx/shape/lib/shapes/Pie'
import { noop } from 'rxjs'
import { PieChartContent } from 'sourcegraph'

import { LockedChart } from '../locked/LockedChart'
import { MaybeLink } from '../MaybeLink'

import { PieArc } from './components/PieArc'
import { distributePieArcs } from './distribute-pie-data'

import styles from './PieChart.module.scss'

// Visual settings
const DEFAULT_FILL_COLOR = 'var(--color-bg-3)'
const DEFAULT_PADDING = { top: 20, right: 20, bottom: 20, left: 20 }

/** Generate percent value for each part of pie chart */
const getSubtitle = <Datum extends object>(arc: PieArcDatum<Datum>, total: number): string =>
    `${((100 * arc.value) / total).toFixed(2)}%`

export interface PieChartProps<Datum extends object> extends PieChartContent<Datum> {
    /** Chart width in px */
    width: number
    /** Chart height in px */
    height: number
    /** Click handler for pie arc-link element. */
    onDatumLinkClick?: (event: React.MouseEvent) => void
    /** Chart padding in px */
    padding?: typeof DEFAULT_PADDING
    locked?: boolean
}

export function PieChart<Datum extends object>(props: PieChartProps<Datum>): ReactElement | null {
    const { width, height, padding = DEFAULT_PADDING, pies, onDatumLinkClick = noop, locked = false } = props

    // We have to track which arc is hovered to change order of rendering.
    // Due the fact svg elements don't have css z-index (in svg only order of renderings matters)
    // we have to render PieArcs in different order to prevent visual label overlapping on small
    // datasets. When user hovers one pie arc we change arcs order in a way to put this
    // hovered arc last in arcs array. By that we sort of change z-index (ordering) of svg element
    // and put hovered label and arc over other arc elements
    const [hoveredArc, setHoveredArc] = useState<PieArcDatum<Datum> | null>(null)

    // For now we can ignore all other pies, we need to render only one pie per chart
    const content = pies[0]
    const { data, dataKey, nameKey, linkURLKey, fillKey } = content

    const sortedData = useMemo(() => distributePieArcs(data, dataKey), [data, dataKey])

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
                (sum, datum) => sum + +datum[dataKey],
                0
            ),
        [sortedData, dataKey]
    )

    if (locked) {
        return <LockedChart className={styles.lockedChart} />
    }

    // Potential problem, we use title/name of pie arc as key, that's not 100% unique value
    // TODO change this we will have id for each pie
    // Because of nature of PieChartProps<D> we have to cast fields from datum
    // cause that's too generic to derive type by ts
    const getKey = (arc: PieArcDatum<Datum>): string => (arc.data[nameKey] as unknown) as string
    const getFillColor = (arc: PieArcDatum<Datum>): string =>
        ((arc.data[fillKey as keyof Datum] as unknown) as string) ?? DEFAULT_FILL_COLOR
    const getLink = (arc: PieArcDatum<Datum>): string | undefined =>
        linkURLKey ? ((arc.data[linkURLKey] as unknown) as string) : undefined

    const getValue = (data: Datum): number => +data[dataKey]

    if (width < 10) {
        return null
    }

    return (
        <svg aria-label="Pie chart" width={width} height={height}>
            <Group top={centerY + padding.top} left={centerX + padding.left}>
                <Pie
                    data={sortedData}
                    pieValue={getValue}
                    outerRadius={radius}
                    cornerRadius={3}
                    pieSort={null}
                    pieSortValues={null}
                    padRadius={40}
                >
                    {pie => {
                        const arcs = hoveredArc
                            ? [...pie.arcs.filter(arc => arc.index !== hoveredArc?.index), hoveredArc]
                            : pie.arcs

                        return (
                            <Group>
                                {arcs.map((arc, index) => (
                                    <MaybeLink
                                        key={getKey(arc)}
                                        to={getLink(arc)}
                                        target="_blank"
                                        rel="noopener"
                                        className={styles.link}
                                        role={getLink(arc) ? 'link' : 'graphics-dataunit'}
                                        aria-label={`Element ${index + 1} of ${arcs.length}. Name: ${getKey(
                                            arc
                                        )}. Value: ${getSubtitle(arc, total)}.`}
                                        onClick={onDatumLinkClick}
                                    >
                                        <PieArc
                                            arc={arc}
                                            path={pie.path}
                                            getColor={getFillColor}
                                            title={getKey(arc)}
                                            subtitle={getSubtitle(arc, total)}
                                            getLink={getLink}
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
