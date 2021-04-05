import React, { MouseEvent, PointerEventHandler, ReactElement, useCallback, useMemo, useState } from 'react'
import classnames from 'classnames'
import { PieChartContent } from 'sourcegraph'
import Pie, { PieArcDatum } from '@visx/shape/lib/shapes/Pie'
import { Group } from '@visx/group'
import { Annotation, Connector } from '@visx/annotation'
import { Arc as ArcType } from 'd3-shape'

// Replace import below on standard @visx/annotation package
// when ticket about bad label positioning will be resolve
// https://github.com/airbnb/visx/issues/1126
import { Label } from '../../annotation/Label'
import { onDatumClick } from '../types'
import { distributePieArcs } from './distribute-pie-data'

const DEFAULT_FILL_COLOR = 'var(--color-bg-3)'
const DEFAULT_MARGIN = { top: 20, right: 20, bottom: 20, left: 20 }
const CONNECTION_LINE_LENGTH = 15
const CONNECTION_LINE_MARGIN = 2
const LABEL_PADDING = 4

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

    // Accessors
    const usage = useCallback((data: Datum): number => +data[dataKey], [dataKey])

    // Potential problem, we use title/name of pie arc as key, that's not 100% unique value
    // TODO change this we will have id for each pie
    const getKey = useCallback(
        // Because of nature of PieChartProps<D> we have to cast fields from datum
        // cause that's too generic to derive type by ts
        (arc: PieArcDatum<Datum>): string => (arc.data[nameKey] as unknown) as string,
        [nameKey]
    )

    const getFillColor = useCallback(
        (arc: PieArcDatum<Datum>): string => ((arc.data[fillKey] as unknown) as string) ?? DEFAULT_FILL_COLOR,
        [fillKey]
    )

    const getLink = useCallback((arc: PieArcDatum<Datum>): string => (arc.data[linkURLKey] as unknown) as string, [
        linkURLKey,
    ])

    if (width < 10) {
        return null
    }

    return (
        <svg className="pie-chart" width={width} height={height}>
            <Group top={centerY + margin.top} left={centerX + margin.left}>
                <Pie
                    data={sortedData}
                    pieValue={usage}
                    outerRadius={radius}
                    cornerRadius={3}
                    pieSort={null}
                    pieSortValues={null}
                    padRadius={40}
                >
                    {pie => (
                        <Group>
                            {pie.arcs.map(arc => (
                                <PieArc
                                    key={getKey(arc)}
                                    visible={getKey(arc) === (activeArc && getKey(activeArc))}
                                    arc={arc}
                                    path={pie.path}
                                    total={total}
                                    getColor={getFillColor}
                                    getKey={getKey}
                                    getLink={getLink}
                                    /* eslint-disable-next-line react/jsx-no-bind */
                                    onPointerMove={() => setActiveArc(arc)}
                                    onClick={onDatumClick}
                                />
                            ))}

                            {/*
                                Due the fact svg elements don't have css z-index (in svg only order of renderings matters)
                                we have to render PieArcs twice to prevent visual label overlapping on small datasets.
                                When user hovers one pie arc above we change the opacity and replace this arc with one
                                of the arcs below by that we sort of change z-index and svg element and put hovered
                                label and arc over other elements above
                             */}
                            {pie.arcs.map(arc => (
                                <PieArc
                                    key={getKey(arc)}
                                    visible={!(getKey(arc) === (activeArc && getKey(activeArc)))}
                                    arc={arc}
                                    path={pie.path}
                                    total={total}
                                    getColor={getFillColor}
                                    getKey={getKey}
                                    getLink={getLink}
                                    /* eslint-disable-next-line react/jsx-no-bind */
                                    onPointerMove={() => setActiveArc(arc)}
                                    onClick={onDatumClick}
                                />
                            ))}
                        </Group>
                    )}
                </Pie>
            </Group>
        </svg>
    )
}

interface PieArcProps<Datum> {
    visible: boolean
    getKey: (d: PieArcDatum<Datum>) => string
    getColor: (d: PieArcDatum<Datum>) => string
    getLink: (d: PieArcDatum<Datum>) => string
    path: ArcType<unknown, PieArcDatum<Datum>>
    arc: PieArcDatum<Datum>
    onClick: onDatumClick
    total: number
    onPointerMove?: PointerEventHandler
    onPointerOut?: PointerEventHandler
}

// Don't have another way to add our own classes to label elements
const TITLE_PROPS = { className: 'pie-chart__label-title' }
const SUBTITLE_PROPS = { className: 'pie-chart__label-sub-title' }

function PieArc<Datum>(props: PieArcProps<Datum>): ReactElement {
    const { visible, total, path, arc, getColor, getKey, getLink, onClick, onPointerMove, onPointerOut } = props

    const pathValue = path(arc) ?? ''
    const name = getKey(arc)
    const link = getLink(arc)
    const labelSubtitle = `${((100 * arc.value) / total).toFixed(2)}%`

    // Math to put label and connection line in a middle of arc radius surface
    const middleAngle = Math.PI / 2 - (arc.startAngle + (arc.endAngle - arc.startAngle) / 2)
    const radius = path.outerRadius()(arc)
    const normalX = Math.cos(middleAngle)
    const normalY = Math.sin(-middleAngle)

    const labelX = normalX * CONNECTION_LINE_LENGTH
    const labelY = normalY * CONNECTION_LINE_LENGTH

    const surfaceX = normalX * (radius + CONNECTION_LINE_MARGIN)
    const surfaceY = normalY * (radius + CONNECTION_LINE_MARGIN)

    // Handlers
    const handleGroupClick = useCallback((event: MouseEvent) => onClick({ originEvent: event, link }), [link, onClick])

    const classes = classnames('pie-chart__arc', {
        'pie-chart__arc--with-link': link,
        'pie-chart__arc--active': visible,
    })

    return (
        <Group className={classes} onClick={handleGroupClick} onPointerMove={onPointerMove} onPointerOut={onPointerOut}>
            <path className="pie-chart__arc-path" d={pathValue} fill={getColor(arc)} />

            <Annotation x={surfaceX} y={surfaceY} dx={labelX} dy={labelY}>
                <Connector className="pie-chart__label-line" type="line" />

                <Label
                    className="pie-chart__label"
                    backgroundPadding={LABEL_PADDING}
                    showBackground={true}
                    showAnchorLine={false}
                    title={name}
                    subtitleDy={0}
                    titleProps={TITLE_PROPS}
                    subtitleProps={SUBTITLE_PROPS}
                    subtitle={labelSubtitle}
                />
            </Annotation>

            <circle className="pie-chart__label-circle" r={2} cx={surfaceX + labelX} cy={surfaceY + labelY} />
        </Group>
    )
}
