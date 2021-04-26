import { Annotation, Connector } from '@visx/annotation'
import { Group } from '@visx/group'
import { PieArcDatum } from '@visx/shape/lib/shapes/Pie'
import classnames from 'classnames'
import { Arc as ArcType } from 'd3-shape'
import React, { PointerEventHandler, ReactElement } from 'react'

import { Label } from '../../../annotation/Label'

// Pie arc visual settings
const CONNECTION_LINE_LENGTH = 15
const CONNECTION_LINE_MARGIN = 2
const LABEL_PADDING = 4

// We have no other way but to add our own classes to label in this way
const TITLE_PROPS = { className: 'pie-chart__label-title' }
const SUBTITLE_PROPS = { className: 'pie-chart__label-sub-title' }

interface PieArcProps<Datum> {
    /** Title for current pie arc */
    title: string
    /** Sub-title (percent value) for current pie arc*/
    subtitle: string
    /** Getter (accessor) to have a color for current pie arc */
    getColor: (d: PieArcDatum<Datum>) => string
    /** Getter (accessor) to have a link for current pie arc */
    getLink: (d: PieArcDatum<Datum>) => string | undefined
    /** The arc generator produces a circular or annular sector, as in a pie or donut chart. */
    path: ArcType<unknown, PieArcDatum<Datum>>
    /** Element of the Arc. The generic refers to the data type of an element in the input array. */
    arc: PieArcDatum<Datum>
    /** Callback to handle pointer (mouse, touch) move over arc */
    onPointerMove?: PointerEventHandler
    /** Callback to handle pointer (mouse, touch) out over arc */
    onPointerOut?: PointerEventHandler
}

/**
 * Display particular arc and annotation for PieChart.
 * */
export function PieArc<Datum>(props: PieArcProps<Datum>): ReactElement {
    const { title, subtitle, path, arc, getColor, getLink, onPointerMove, onPointerOut } = props

    const pathValue = path(arc) ?? ''
    const link = getLink(arc)

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

    const classes = classnames('pie-chart__arc', {
        'pie-chart__arc--with-link': link,
    })

    return (
        <Group aria-hidden={true} className={classes} onPointerMove={onPointerMove} onPointerOut={onPointerOut}>
            <path className="pie-chart__arc-path" d={pathValue} fill={getColor(arc)} />

            <Annotation x={surfaceX} y={surfaceY} dx={labelX} dy={labelY}>
                <Connector className="pie-chart__label-line" type="line" />

                <Label
                    className="pie-chart__label"
                    backgroundPadding={LABEL_PADDING}
                    showBackground={true}
                    showAnchorLine={false}
                    title={title}
                    subtitleDy={0}
                    titleFontWeight={200}
                    subtitleFontWeight={200}
                    titleProps={TITLE_PROPS}
                    subtitleProps={SUBTITLE_PROPS}
                    subtitle={subtitle}
                />
            </Annotation>

            <circle className="pie-chart__label-circle" r={2} cx={surfaceX + labelX} cy={surfaceY + labelY} />
        </Group>
    )
}
