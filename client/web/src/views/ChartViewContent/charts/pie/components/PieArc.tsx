import React, { MouseEvent, PointerEventHandler, ReactElement, useCallback } from 'react'
import { PieArcDatum } from '@visx/shape/lib/shapes/Pie'
import { Arc as ArcType } from 'd3-shape'
import classnames from 'classnames'
import { Group } from '@visx/group'
import { Annotation, Connector } from '@visx/annotation'

// Replace import below on standard @visx/annotation package
// when ticket about bad label positioning will be resolve
// https://github.com/airbnb/visx/issues/1126
import { Label } from '../../../annotation/Label'
import { onDatumClick } from '../../types'

interface PieArcProps<Datum> {
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

const CONNECTION_LINE_LENGTH = 15
const CONNECTION_LINE_MARGIN = 2
const LABEL_PADDING = 4

// Don't have another way to add our own classes to label elements
const TITLE_PROPS = { className: 'pie-chart__label-title' }
const SUBTITLE_PROPS = { className: 'pie-chart__label-sub-title' }

export function PieArc<Datum>(props: PieArcProps<Datum>): ReactElement {
    const { total, path, arc, getColor, getKey, getLink, onClick, onPointerMove, onPointerOut } = props

    const pathValue = path(arc) ?? ''
    const name = getKey(arc)
    const link = getLink(arc)
    const labelSubtitle = `${((100 * arc.value) / total).toFixed(2)}%`

    // Math to put label and connection line in a middle of arc radius surface
    // Find the middle of the arc segment. Here we have polar system of coordinate.
    // In polar system we have angle and radius to describe the point.
    const middleAngle = Math.PI / 2 - (arc.startAngle + (arc.endAngle - arc.startAngle) / 2)
    const radius = path.outerRadius()(arc)

    // Pie chart operate polar system of coords but svg operates with a Cartesian coordinate system
    // normalX and normalY they are just projections on the axes. You can thing about this code like
    // transformation polar coords to cartesian coords.
    // https://en.wikipedia.org/wiki/Polar_coordinate_system#/media/File:Polar_to_cartesian.svg
    const normalX = Math.cos(middleAngle)
    const normalY = Math.sin(-middleAngle)

    // Calculate coords for start point of label line
    const surfaceX = normalX * (radius + CONNECTION_LINE_MARGIN)
    const surfaceY = normalY * (radius + CONNECTION_LINE_MARGIN)

    // Calculate coords for end point of label line
    const labelX = normalX * CONNECTION_LINE_LENGTH
    const labelY = normalY * CONNECTION_LINE_LENGTH

    // Handlers
    const handleGroupClick = useCallback((event: MouseEvent) => onClick({ originEvent: event, link }), [link, onClick])

    const classes = classnames('pie-chart__arc', {
        'pie-chart__arc--with-link': link,
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
