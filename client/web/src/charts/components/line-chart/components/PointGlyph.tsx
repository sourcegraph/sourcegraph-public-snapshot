import { GlyphDot } from '@visx/glyph'
import React, { FocusEventHandler } from 'react'

import { MaybeLink } from '../../../../views/components/view/content/chart-view-content/charts/MaybeLink'

const stopPropagation = (event: React.MouseEvent): void => event.stopPropagation()

interface PointGlyphProps {
    top: number
    left: number
    color: string
    active: boolean
    linkURL?: string
    onFocus: FocusEventHandler<Element>
    onBlur: FocusEventHandler<Element>
}

export const PointGlyph: React.FunctionComponent<PointGlyphProps> = props => {
    const { top, left, color, active, linkURL, onFocus, onBlur } = props

    return (
        <MaybeLink
            to={linkURL}
            target="_blank"
            rel="noopener"
            onPointerUp={stopPropagation}
            onClick={stopPropagation}
            onFocus={onFocus}
            onBlur={onBlur}
            role={linkURL ? 'link' : 'graphics-dataunit'}
        >
            <GlyphDot
                tabIndex={linkURL ? -1 : 0}
                onFocus={onFocus}
                onBlur={onBlur}
                cx={left}
                cy={top}
                stroke={color}
                fill="var(--body-bg)"
                strokeWidth={active ? 3 : 2}
                r={active ? 6 : 4}
            />
        </MaybeLink>
    )
}
