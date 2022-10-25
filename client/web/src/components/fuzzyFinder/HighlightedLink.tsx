import React from 'react'

import classNames from 'classnames'

import { Link, Code } from '@sourcegraph/wildcard'

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
                <mark key={key} className={classNames(styles.mark, 'px-0')}>
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

    return props.url ? (
        <Code>
            <Link key="link" tabIndex={-1} className={styles.link} to={props.url} onClick={props.onClick}>
                {props.icon && <span key="icon">{props.icon}</span>}
                {spans}
                {props.textSuffix}
            </Link>
        </Code>
    ) : (
        <Link
            key="link"
            tabIndex={-1}
            className={styles.link}
            to={`/commands/${props.text}`}
            onClick={event => {
                event.preventDefault()
                props.onClick?.()
            }}
        >
            {props.icon && <span key="icon">{props.icon}</span>}
            {spans}
        </Link>
    )
}

export const linkStyle = styles.link
