import React, { ReactElement, useMemo, useState } from 'react'
import { PieChartContent } from 'sourcegraph'
import Pie, { PieArcDatum } from '@visx/shape/lib/shapes/Pie'
import { Group } from '@visx/group'

import { onDatumClick } from '../types'
import { distributePieArcs } from './distribute-pie-data'
import { PieArc } from './components/PieArc'

const DEFAULT_FILL_COLOR = 'var(--color-bg-3)'
const DEFAULT_MARGIN = { top: 20, right: 20, bottom: 20, left: 20 }

export interface PieChartProps<Datum extends object> extends PieChartContent<Datum> {
    width: number
    height: number
    margin?: typeof DEFAULT_MARGIN
    onDatumClick: onDatumClick
}

export function PieChart<Datum extends object>(props: PieChartProps<Datum>): ReactElement | null {
    const { width, height, margin = DEFAULT_MARGIN, pies, onDatumClick } = props

    const [activeArc, setActiveArc] = useState<PieArcDatum<Datum> | null>(null)

    // For now we can ignore all other pies, we need to render only one pie per chart
    const content = pies[0]
    const { data, dataKey, nameKey, linkURLKey = '' as keyof Datum, fillKey = '' as keyof Datum } = content

    const sortedData = useMemo(() => distributePieArcs(data, dataKey), [data, dataKey])

    const innerWidth = width - margin.left - margin.right
    const innerHeight = height - margin.top - margin.bottom

    const radius = Math.min(innerWidth, innerHeight) / 3
    const centerY = innerHeight / 2
    const centerX = innerWidth / 2

    const total = useMemo(
        () =>
            sortedData.reduce(
                // Here we have to cast datum[dataKey] to number because we ts can't derive value by dataKey
                (sum, datum) => sum + +datum[dataKey],
                0
            ),
        [sortedData, dataKey]
    )

    // Potential problem, we use title/name of pie arc as key, that's not 100% unique value
    // TODO change this we will have id for each pie
    // Because of nature of PieChartProps<D> we have to cast fields from datum
    // cause that's too generic to derive type by ts
    const getKey = (arc: PieArcDatum<Datum>): string => (arc.data[nameKey] as unknown) as string
    const getFillColor = (arc: PieArcDatum<Datum>): string =>
        ((arc.data[fillKey] as unknown) as string) ?? DEFAULT_FILL_COLOR
    const getLink = (arc: PieArcDatum<Datum>): string => (arc.data[linkURLKey] as unknown) as string

    // Accessors
    const getValue = (data: Datum): number => +data[dataKey]

    if (width < 10) {
        return null
    }

    return (
        /* eslint-disable react/jsx-no-bind */
        <svg className="pie-chart" width={width} height={height}>
            <Group top={centerY + margin.top} left={centerX + margin.left}>
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
                        const arcs = activeArc
                            ? [...pie.arcs.filter(arc => arc.index !== activeArc?.index), activeArc]
                            : pie.arcs

                        return (
                            <Group>
                                {arcs.map(arc => (
                                    <PieArc
                                        key={getKey(arc)}
                                        arc={arc}
                                        path={pie.path}
                                        total={total}
                                        getColor={getFillColor}
                                        getKey={getKey}
                                        getLink={getLink}
                                        onPointerMove={() => setActiveArc(arc)}
                                        onClick={onDatumClick}
                                    />
                                ))}
                            </Group>
                        )
                    }}
                </Pie>
            </Group>
        </svg>
    )
}
