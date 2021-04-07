import React, { ReactElement } from 'react'
import { GlyphProps } from '@visx/xychart'
import { GlyphDot as Glyph } from '@visx/glyph'

export function GlyphDot<Datum extends object>(props: GlyphProps<Datum>): ReactElement {
    const { x: xCoord, y: yCoord, color, onPointerMove, onPointerOut, onPointerUp } = props
    const handlers = { onPointerMove, onPointerOut, onPointerUp }

    return <Glyph className="line-chart__glyph" cx={xCoord} cy={yCoord} fill={color} r={6} {...handlers} />
}
