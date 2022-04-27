import { MouseEventHandler, PointerEventHandler, ReactElement } from 'react'

import { GlyphDot as Glyph } from '@visx/glyph'
import { EventHandlerParams, GlyphProps } from '@visx/xychart/lib/types'
import classNames from 'classnames'
import { LineChartSeries } from 'sourcegraph'

import { MaybeLink } from '../../MaybeLink'
import { Point } from '../types'

import { getLineStroke } from './LineChartContent'
import { dateLabelFormatter } from './TickComponent'

import styles from './GlyphContent.module.scss'

/**
 * Type for active datum state in LineChartContent component. In order to render active state
 * for hovered or focused point we need to track active datum to calculate styles for active glyph.
 */
export interface ActiveDatum<Datum extends object> extends EventHandlerParams<Point> {
    /** Series of data of active datum */
    line?: LineChartSeries<Datum>
}

interface GlyphContentProps<Datum extends object> extends Omit<GlyphProps<Point>, 'key' | 'index'> {
    /**
     * Just because GlyphProps has a bug with types.
     * GlyphProps key is an index of current datum and
     * GlyphProps index doesn't exist in runtime.
     */
    index: string

    /** Hovered point info (datum) to calculate proper styles for particular Glyph */
    hoveredDatum: ActiveDatum<Datum> | null

    /** Focused point info (datum) to calculate proper styles for particular Glyph */
    focusedDatum: ActiveDatum<Datum> | null

    /** Line (series) index of current point */
    lineIndex: number

    /** Total number of lines (series) to calculate proper aria-label for glyph content */
    totalNumberOfLines: number

    /** Data of particular line of current glyph (chart point) */
    line: LineChartSeries<Datum>

    /** Focus handler for glyph (chart point) */
    setFocusedDatum: (datum: ActiveDatum<Datum> | null) => void

    /** On click handler for root component of glyph content */
    onClick: MouseEventHandler

    /** On pointer up handler for root component of glyph content */
    onPointerUp: PointerEventHandler
}

/** Displays glyph (chart point) with link */
export function GlyphContent<Datum extends object>(props: GlyphContentProps<Datum>): ReactElement {
    const {
        index,
        line,
        hoveredDatum,
        focusedDatum,
        datum,
        lineIndex,
        totalNumberOfLines,
        x: xCoordinate,
        y: yCoordinate,
        onPointerUp,
        onClick,
        setFocusedDatum,
    } = props

    const currentDatumIndex = +index
    const hovered = hoveredDatum?.index === currentDatumIndex && hoveredDatum.key === line.dataKey
    const focused = focusedDatum?.index === currentDatumIndex && focusedDatum.key === line.dataKey

    const linkURL = line.linkURLs?.[+datum.x] ?? line.linkURLs?.[currentDatumIndex]

    const currentDatum = {
        key: line.dataKey.toString(),
        index: currentDatumIndex,
        datum,
    }

    const xAxisValue = dateLabelFormatter(new Date(datum.x))
    const yAxisValue = ((datum.y as unknown) as string) ?? ''
    const ariaLabel = `Point ${currentDatumIndex + 1} of line ${
        lineIndex + 1
    } of ${totalNumberOfLines}. X value: ${xAxisValue}. Y value: ${yAxisValue}`

    return (
        <MaybeLink
            to={linkURL}
            target="_blank"
            rel="noopener"
            onPointerUp={onPointerUp}
            onClick={onClick}
            onFocus={() => linkURL && setFocusedDatum(currentDatum)}
            onBlur={() => linkURL && setFocusedDatum(null)}
            className={styles.glyphLink}
            role={linkURL ? 'link' : 'graphics-dataunit'}
            aria-label={ariaLabel}
        >
            <Glyph
                className={classNames(styles.glyph, hovered && styles.glyphActive)}
                cx={xCoordinate}
                cy={yCoordinate}
                stroke={getLineStroke(line)}
                r={hovered || focused ? 6 : 4}
            />
        </MaybeLink>
    )
}
