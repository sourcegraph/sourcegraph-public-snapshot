import React from 'react'
import { Link } from 'react-router-dom'

import styles from './HighlightedText.module.scss'

export interface RangePosition {
    startOffset: number
    endOffset: number
    /**
     * Does this range enclose an exact word?
     */
    isExact: boolean
}
export interface HighlightedTextProps {
    text: string
    positions: RangePosition[]
    url?: string
}

export function offsetSum(props: HighlightedTextProps): number {
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
export const HighlightedText: React.FunctionComponent<HighlightedTextProps> = props => {
    const spans: JSX.Element[] = []
    let start = 0
    function pushSpan(className: string, startOffset: number, endOffset: number): void {
        if (startOffset >= endOffset) {
            return
        }
        const text = props.text.slice(startOffset, endOffset)
        const key = `${startOffset}-${endOffset}`
        const span = (
            <span key={key} className={className}>
                {text}
            </span>
        )
        spans.push(span)
    }
    for (const position of props.positions) {
        if (position.startOffset > start) {
            pushSpan('', start, position.startOffset)
        }
        start = position.endOffset
        const classNameSuffix = position.isExact ? styles.exact : styles.fuzzy
        pushSpan(`${styles.highlighted} ${classNameSuffix}`, position.startOffset, position.endOffset)
    }
    pushSpan('', start, props.text.length)

    return props.url ? (
        <Link className={styles.link} to={props.url}>
            {spans}
        </Link>
    ) : (
        <>{spans}</>
    )
}
