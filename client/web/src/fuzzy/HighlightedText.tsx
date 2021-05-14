import React from 'react'

export interface RangePosition {
    startOffset: number
    endOffset: number
    /**
     * Does this range enclose an exact word?
     */
    isExact: boolean
}

export class HighlightedTextProps {
    constructor(
        public readonly text: string,
        public readonly positions: RangePosition[],
        public readonly url?: string
    ) {}
    public offsetSum(): number {
        let sum = 0
        for (const position of this.positions) {
            sum += position.startOffset
        }
        return sum
    }
    public exactCount(): number {
        let result = 0
        for (const position of this.positions) {
            if (position.isExact) {
                result++
            }
        }
        return result
    }
    public isExact(): boolean {
        return this.positions.length === 1 && this.positions[0].isExact
    }
}

export interface HighlightedTextPropsInstance {
    value: HighlightedTextProps
}

/**
 * React component that re
 */
export const HighlightedText: React.FunctionComponent<HighlightedTextPropsInstance> = propsInstance => {
    const props = propsInstance.value
    const spans: JSX.Element[] = []
    let start = 0
    function pushSpan(className: string, startOffset: number, endOffset: number): void {
        if (startOffset >= endOffset) {
            return
        }
        const text = props.text.slice(startOffset, endOffset)
        const key = `${text}-${className}`
        const span = (
            <span key={key} className={className}>
                {text}
            </span>
        )
        spans.push(span)
    }
    for (const position of props.positions) {
        if (position.startOffset > start) {
            pushSpan('fuzzy-modal-plaintext', start, position.startOffset)
        }
        start = position.endOffset
        const classNameSuffix = position.isExact ? 'exact' : 'fuzzy'
        pushSpan(
            `fuzzy-modal-highlighted fuzzy-modal-highlighted-${classNameSuffix}`,
            position.startOffset,
            position.endOffset
        )
    }
    pushSpan('fuzzy-modal-plaintext', start, props.text.length)

    return props.url ? (
        <a className="fuzzy-modal-link" href={props.url}>
            {spans}
        </a>
    ) : (
        <>{spans}</>
    )
}
