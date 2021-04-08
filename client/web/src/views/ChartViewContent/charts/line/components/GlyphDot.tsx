import React, { ReactElement } from 'react'
import { GlyphProps } from '@visx/xychart'
import { GlyphDot as Glyph } from '@visx/glyph'

/**
 * Displays glyph (point) on the chart with our own className so that we can style glyphs by css.
 * */
export function GlyphDot<Datum extends object>(props: GlyphProps<Datum>): ReactElement {
    const { x: xCoord, y: yCoord, color, onPointerMove, onPointerOut, onPointerUp } = props
    const handlers = { onPointerMove, onPointerOut, onPointerUp }

    return <Glyph className="line-chart__glyph" cx={xCoord} cy={yCoord} stroke={color} r={4} {...handlers} />
}
