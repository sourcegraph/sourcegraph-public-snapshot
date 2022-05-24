import React from 'react'

import { ContentMatch } from '@sourcegraph/shared/src/search/stream'

import styles from './FileSearchResult.module.scss'

interface Props {
    line: ContentMatch['lineMatches'][0]
}

export const TrimmedCodeLineWithHighlights: React.FunctionComponent<Props> = React.memo<Props>(({ line }) => {
    const content = line.line
    const trimLeft = content.length - content.trimStart().length

    // We create a mutable copy of the array here, so we can later remove elements from it to
    // remember our progress.
    const highlightRanges = [...line.offsetAndLengths]

    const segments = []
    for (let index = trimLeft; index <= content.length; ) {
        const nextPointOfInterest = highlightRanges.shift()

        // There are no more highlight segments, so we can render the remaining text
        if (nextPointOfInterest === undefined) {
            segments.push(content.slice(index))
            break
        } else {
            // There are highlight segments, so we need to render the text before the next highlight
            // and then render the highlight.
            const [offset, length] = nextPointOfInterest
            if (content.slice(index, offset) !== '') {
                segments.push(content.slice(index, offset))
            }
            segments.push(
                <span key={`highlight-${index}`} className={styles.lineCodeHighlight}>
                    {content.slice(offset, offset + length)}
                </span>
            )
            index = offset + length
        }
    }

    return <>{segments}</>
})
