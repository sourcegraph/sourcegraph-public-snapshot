import React, { useCallback } from 'react'

import classNames from 'classnames'

import { Link } from '@sourcegraph/wildcard'

import styles from './HighlightedLink.module.scss'

export interface RangePosition {
    startOffset: number
    endOffset: number
    /**
     * Does this range enclose an exact word?
     */
    isExact: boolean
}
export interface HighlightedLinkProps {
    text: string
    positions: RangePosition[]
    url?: string
    icon?: JSX.Element
    textSuffix?: JSX.Element
    onClick?: () => void
    // Fuzzy finding score, used to sort aggregated results across different
    // fuzzy finder tabs.
    score?: number
}

export function offsetSum(props: HighlightedLinkProps): number {
    let sum = 0
    for (const position of props.positions) {
        sum += position.startOffset
    }
    return sum
}

/**
 * React component that renders text with highlighted subranges.
 *
 * Used to render fuzzy finder results. For example, given the query "doc/read"
 * we want to highlight 'Doc' and `READ' in the filename
 * 'Documentation/README.md`.
 */
export const HighlightedLink: React.FunctionComponent<React.PropsWithChildren<HighlightedLinkProps>> = props => {
    const spans: JSX.Element[] = []
    let start = 0
    function pushElement(kind: 'mark' | 'span', startOffset: number, endOffset: number): void {
        if (startOffset >= endOffset) {
            return
        }
        const text = props.text.slice(startOffset, endOffset)
        const key = `${startOffset}-${endOffset}`
        if (kind === 'mark') {
            spans.push(
                <mark key={key} className="px-0">
                    {text}
                </mark>
            )
        } else {
            spans.push(<span key={key}>{text}</span>)
        }
    }
    for (const [index, position] of props.positions.entries()) {
        if (index > 0) {
            const previous = props.positions[index - 1]
            if (
                previous.startOffset === position.startOffset &&
                previous.endOffset === position.endOffset &&
                previous.isExact === position.isExact
            ) {
                continue
            }
        }
        if (position.startOffset > start) {
            pushElement('span', start, position.startOffset)
        }
        start = position.endOffset
        pushElement('mark', position.startOffset, position.endOffset)
    }
    pushElement('span', start, props.text.length)

    const { url, onClick } = props
    const handleClick = useCallback(
        (event: React.MouseEvent<HTMLAnchorElement, MouseEvent>) => {
            if (!url) {
                event.preventDefault()
            }
            onClick?.()
        },
        [url, onClick]
    )

    return (
        <Link
            className={classNames('d-inline-block w-100 h-100 text-decoration-none', styles.link)}
            to={url || `/commands/${props.text}`}
            onClick={handleClick}
        >
            {props.icon && <span>{props.icon}</span>}
            {spans}
            {url ? props.textSuffix : null}
        </Link>
    )
}

export const linkStyle = styles.link
