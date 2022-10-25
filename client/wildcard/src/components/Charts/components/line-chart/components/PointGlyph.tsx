import React, { FocusEventHandler, MouseEventHandler } from 'react'

import { GlyphDot } from '@visx/glyph'

import { MaybeLink } from '../../../core'

interface PointGlyphProps {
    top: number
    left: number
    color: string
    active: boolean
    linkURL?: string
    tabIndex?: number
    onClick: MouseEventHandler<Element>
    onFocus?: FocusEventHandler<Element>
    onBlur?: FocusEventHandler<Element>
}

export const PointGlyph: React.FunctionComponent<React.PropsWithChildren<PointGlyphProps>> = props => {
    const { top, left, color, active, linkURL, tabIndex = 0, onFocus, onBlur, onClick } = props

    return (
        <MaybeLink
            to={linkURL}
            target="_blank"
            rel="noopener"
            onClick={onClick}
            onFocus={onFocus}
            onBlur={onBlur}
            tabIndex={tabIndex}
            role={linkURL ? 'link' : 'graphics-dataunit'}
            aria-label={linkURL ? 'Click to view data point detail' : 'Data point'}
        >
            <GlyphDot
                tabIndex={linkURL ? -1 : tabIndex}
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
