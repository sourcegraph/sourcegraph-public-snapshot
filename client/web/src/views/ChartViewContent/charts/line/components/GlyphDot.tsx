import React, { ReactElement } from 'react';
import { GlyphProps } from '@visx/xychart'
import { GlyphDot } from '@visx/glyph'

export function GlyphDotComponent(props: GlyphProps<any>): ReactElement {
    const { x: xCoord, y: yCoord, color, onPointerMove, onPointerOut, onPointerUp } = props;
    const handlers = { onPointerMove, onPointerOut, onPointerUp };

    return (
        <GlyphDot
            className='line-chart__glyph'
            cx={xCoord}
            cy={yCoord}
            fill={color}
            r={6}
            {...handlers}
        />
    );
}
