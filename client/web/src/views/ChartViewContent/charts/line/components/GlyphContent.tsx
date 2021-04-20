import { GlyphDot as Glyph } from '@visx/glyph'
import { EventHandlerParams, GlyphProps } from '@visx/xychart/lib/types'
import classnames from 'classnames'
import React, { MouseEventHandler, PointerEventHandler, ReactElement } from 'react'

import { MaybeLink } from '../../MaybeLink'
import { Accessors } from '../types'

import { getLineStroke, LineChartContentProps } from './LineChartContent'
import { dateLabelFormatter } from './TickComponent'

/**
 * Type for active datum state in LineChartContent component. In order to render active state
 * for hovered or focused point we need to track active datum to calculate styles for active glyph.
 */
export interface ActiveDatum<Datum extends object> extends EventHandlerParams<Datum> {
    /** Series of data of active datum */
    line?: LineChartContentProps<Datum>['series'][number]
}

interface GlyphContentProps<Datum extends object> extends Omit<GlyphProps<Datum>, 'key' | 'index'> {
    /**
     * Just because GlyphProps has a bug with types.
     * GlyphProps key is an index of current datum and
     * GlyphProps index doesn't exist in runtime.
     * */
    index: string
    /** Hovered point info (datum) to calculate proper styles for particular Glyph */
    hoveredDatum: ActiveDatum<Datum> | null
    /** Focused point info (datum) to calculate proper styles for particular Glyph */
    focusedDatum: ActiveDatum<Datum> | null
    /** Focus handler for glyph (chart point) */
    setFocusedDatum: (datum: ActiveDatum<Datum> | null) => void
    /** Map with getters to have a proper value of by x and y axis value for current point */
    accessors: Accessors<Datum, keyof Datum>
    /** Line (series) index of current point */
    lineIndex: number
    /** Total number of lines (series) to calculate proper aria-label for glyph content */
    totalNumberOfLines: number
    /** Data of particular line of current glyph (chart point) */
    line: LineChartContentProps<Datum>['series'][number]
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
        accessors,
        lineIndex,
        totalNumberOfLines,
        onPointerUp,
        onClick,
        setFocusedDatum,
        x: xCoordinate,
        y: yCoordinate,
    } = props

    const currentDatumIndex = +index
    const hovered = hoveredDatum?.index === currentDatumIndex && hoveredDatum.key === line.dataKey
    const focused = focusedDatum?.index === currentDatumIndex && focusedDatum.key === line.dataKey

    const linkURL = line.linkURLs?.[currentDatumIndex]
    const currentDatum = {
        key: line.dataKey.toString(),
        index: currentDatumIndex,
        datum,
    }

    const xAxisValue = dateLabelFormatter(new Date(accessors.x(datum)))
    const yAxisValue = (accessors.y?.[line.dataKey](datum) as string) ?? ''
    const ariaLabel = `Point ${currentDatumIndex + 1} of line ${
        lineIndex + 1
    } of ${totalNumberOfLines}. X value: ${xAxisValue}. Y value: ${yAxisValue}`

    return (
        <MaybeLink
            to={linkURL}
            onPointerUp={onPointerUp}
            onClick={onClick}
            /* eslint-disable-next-line react/jsx-no-bind */
            onFocus={() => linkURL && setFocusedDatum(currentDatum)}
            /* eslint-disable-next-line react/jsx-no-bind */
            onBlur={() => linkURL && setFocusedDatum(null)}
            className="line-chart__glyph-link"
            role={linkURL ? 'link' : 'graphics-dataunit'}
            aria-label={ariaLabel}
        >
            <Glyph
                className={classnames('line-chart__glyph', {
                    'line-chart__glyph--active': hovered,
                })}
                cx={xCoordinate}
                cy={yCoordinate}
                stroke={getLineStroke(line)}
                r={hovered || focused ? 6 : 4}
            />
        </MaybeLink>
    )
}
