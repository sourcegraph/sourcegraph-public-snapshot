import * as React from 'react'

/** Wraps matches of pattern in text with <strong>. */
export const HighlightedMatches: React.FunctionComponent<{
    /* The text to display matches in. */
    text: string

    /* The pattern to highlight in the text. */
    pattern: string

    /** The class name for the <strong> element for matches. */
    className?: string
}> = ({ text, pattern, className }) => (
    <span>
        {fuzzyMatches(text.toLowerCase(), pattern.toLowerCase()).map((span, i) =>
            span.match ? (
                <strong key={i} className={className}>
                    {text.slice(span.start, span.end)}
                </strong>
            ) : (
                text.slice(span.start, span.end)
            )
        )}
    </span>
)

export interface Span {
    start: number
    end: number
    match: boolean
}

export function fuzzyMatches(text: string, pattern: string): Span[] {
    const matches: Span[] = []
    let i = 0
    let last: Span | undefined
    for (let pos = 0; pos < pattern.length; pos++) {
        const ti = text.indexOf(pattern.charAt(pos), i)
        if (ti === -1) {
            break
        }
        if (last?.match && ti === i) {
            last.end = ti + 1
        } else if ((last && !last.match) || ti > i) {
            matches.push({ start: i, end: ti, match: false })
            last = undefined
        }
        if (!last) {
            last = { start: ti, end: ti + 1, match: true }
            matches.push(last)
        }
        i = ti + 1
    }
    if (text && i < text.length) {
        matches.push({ start: i, end: text.length, match: false })
    }
    return matches
}
