import type { ReactElement, SVGProps } from 'react'

import { Annotation, Connector, HtmlLabel } from '@visx/annotation'
import { Group } from '@visx/group'
import type { PieArcDatum } from '@visx/shape/lib/shapes/Pie'
import classNames from 'classnames'
import type { Arc as ArcType } from 'd3-shape'

import { Text } from '../../../../Typography'
import { DEFAULT_FALLBACK_COLOR } from '../../../constants'

import styles from './PieArc.module.scss'

// Pie arc visual settings
const CONNECTION_LINE_LENGTH = 15
const CONNECTION_LINE_MARGIN = 2

interface PieArcProps<Datum> extends Omit<SVGProps<SVGGElement>, 'path'> {
    title: string
    subtitle: string
    path: ArcType<unknown, PieArcDatum<Datum>>
    arc: PieArcDatum<Datum>
    className?: string
    getColor: (datum: Datum) => string | undefined
}

export function PieArc<Datum>(props: PieArcProps<Datum>): ReactElement {
    const { title, subtitle, path, arc, getColor, className, ...attributes } = props

    const pathValue = path(arc) ?? ''

    // Math to put label and connection line in a middle of arc radius surface
    // Find the middle of the arc segment. Here we have polar system of coordinate.
    // In polar system we have angle and radius to describe the point.
    const middleAngle = Math.PI / 2 - (arc.startAngle + (arc.endAngle - arc.startAngle) / 2)
    const radius = path.outerRadius()(arc)

    // Pie chart operate polar system of coords but svg operates with a Cartesian coordinate system
    // normalX and normalY they are just projections on the axes. You can think about this code like
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

    return (
        <Group {...attributes} className={styles.arc}>
            <path
                data-testid="pie-chart-arc-element"
                d={pathValue}
                fill={getColor(arc.data) ?? DEFAULT_FALLBACK_COLOR}
                className={classNames(className, styles.arcPath)}
            />

            <Annotation x={surfaceX} y={surfaceY} dx={labelX} dy={labelY}>
                <Connector className={styles.labelLine} type="line" />

                <HtmlLabel showAnchorLine={false} className={styles.label}>
                    <Text className={styles.labelTitle}>{title}</Text>
                    <small className={styles.labelSubTitle}>{subtitle}</small>
                </HtmlLabel>
            </Annotation>

            <circle
                aria-hidden={true}
                className={styles.labelCircle}
                r={2}
                cx={surfaceX + labelX}
                cy={surfaceY + labelY}
            />
        </Group>
    )
}
