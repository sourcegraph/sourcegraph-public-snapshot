import React from 'react'

import classNames from 'classnames'

import { Code, Link } from '@sourcegraph/wildcard'

import styles from './HighlightedLink.module.scss'

export interface RangePosition {
    startOffset: number
    endOffset: number
    /**
     * Does this range enclose an exact word?
     */
    isExact: boolean
}

export const highlightedSectionStyles = {
    muted: styles.mutedSection,
}

export interface HighlightedLinkSection {
    text: string
    positions: RangePosition[]
    style?: keyof typeof highlightedSectionStyles
}
export interface HighlightedLinkProps {
    sections: HighlightedLinkSection[]
    url?: string
    icon?: JSX.Element
    textSuffix?: JSX.Element
    onClick?: () => void
}

export function offsetSum(props: HighlightedLinkProps): number {
    let sum = 0
    for (const section of props.sections) {
        for (const position of section.positions) {
            sum += position.startOffset
        }
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
    function pushElement(
        spans: JSX.Element[],
        section: HighlightedLinkSection,
        kind: 'mark' | 'span',
        startOffset: number,
        endOffset: number
    ): void {
        if (startOffset >= endOffset) {
            return
        }
        const text = section.text.slice(startOffset, endOffset)
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
    const sections: JSX.Element[] = []
    for (const section of props.sections) {
        const spans: JSX.Element[] = []
        let start = 0
        for (const [index, position] of section.positions.entries()) {
            if (index > 0) {
                const previous = section.positions[index - 1]
                if (
                    previous.startOffset === position.startOffset &&
                    previous.endOffset === position.endOffset &&
                    previous.isExact === position.isExact
                ) {
                    continue
                }
            }
            if (position.startOffset > start) {
                pushElement(spans, section, 'span', start, position.startOffset)
            }
            start = position.endOffset
            pushElement(spans, section, 'mark', position.startOffset, position.endOffset)
        }
        pushElement(spans, section, 'span', start, section.text.length)
        sections.push(<span className={section.style ? highlightedSectionStyles[section.style] : ''}>{spans}</span>)
    }

    return props.url ? (
        <Code>
            <Link key="link" tabIndex={-1} className={styles.link} to={props.url} onClick={props.onClick}>
                {props.icon && <span key="icon">{props.icon}</span>}
                {sections}
                {props.textSuffix}
            </Link>
        </Code>
    ) : (
        <Link
            key="link"
            tabIndex={-1}
            className={styles.link}
            to={`/commands/${props.sections.map(({ text }) => text).join('')}`}
            onClick={event => {
                event.preventDefault()
                props.onClick?.()
            }}
        >
            {props.icon && <span key="icon">{props.icon}</span>}
            {sections}
        </Link>
    )
}

export const linkStyle = styles.link
