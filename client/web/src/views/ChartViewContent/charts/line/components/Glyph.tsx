import React, { ReactElement } from 'react';
import { GlyphProps } from '@visx/xychart'
import { GlyphDot } from '@visx/glyph'

export function GlyphComponent(props: GlyphProps<any>): ReactElement {
    const { x: xCoord, y: yCoord, color, onPointerMove, onPointerOut, onPointerUp } = props;
    const handlers = { onPointerMove, onPointerOut, onPointerUp };

    return (
        <GlyphDot
            cx={xCoord}
            cy={yCoord}
            stroke="white"
            strokeWidth={2}
            fill={color}
            r={4}
            {...handlers}
        />
    );
}
