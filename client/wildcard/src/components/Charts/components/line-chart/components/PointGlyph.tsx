import type { FC, FocusEventHandler, MouseEventHandler } from 'react'

import { GlyphDot } from '@visx/glyph'

import { MaybeLink } from '../../../core'

interface PointGlyphProps {
    id: string
    top: number
    left: number
    color: string
    active: boolean
    role: string
    'aria-label': string
    linkURL?: string
    tabIndex?: number
    onClick?: MouseEventHandler<Element>
    onFocus?: FocusEventHandler<Element>
    onBlur?: FocusEventHandler<Element>
}

export const PointGlyph: FC<PointGlyphProps> = props => {
    const {
        id,
        top,
        left,
        color,
        active,
        role,
        'aria-label': ariaLabel,
        linkURL,
        tabIndex = 0,
        onFocus,
        onBlur,
        onClick,
    } = props

    return (
        <MaybeLink
            data-id={id}
            to={linkURL}
            target="_blank"
            rel="noopener"
            tabIndex={tabIndex}
            role={role}
            aria-label={ariaLabel}
            onClickCapture={onClick}
            onFocus={onFocus}
            onBlur={onBlur}
        >
            <GlyphDot
                cx={left}
                cy={top}
                r={active ? 6 : 4}
                fill="var(--body-bg)"
                stroke={color}
                strokeWidth={active ? 3 : 2}
                aria-hidden={true}
                onFocus={onFocus}
                onBlur={onBlur}
            />
        </MaybeLink>
    )
}
