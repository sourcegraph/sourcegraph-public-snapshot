import React from 'react'

import type { LineMatch } from '@sourcegraph/shared/src/search/stream'

import styles from './FileSearchResult.module.scss'

interface Props {
    line: LineMatch
}

export const TrimmedCodeLineWithHighlights: React.FunctionComponent<Props> = React.memo<Props>(
    function TrimmedCodeLineWithHighlights({ line }) {
        const content = line.line
        const trimLeft = content.length - content.trimStart().length

        // We create a mutable copy of the array here, so we can later remove elements from it to
        // remember our progress.
        const highlightRanges = [...line.offsetAndLengths]

        // The indices that we receive from the Sourcegraph API are unicode code pointer offsets
        // rather than byte lengths.
        //
        // We use the spread syntax to properly split a string based into its unicode code points.
        //
        //   "🚀".length      // => 2
        //   [..."🚀"].length // => 1
        const contentArray = [...content]

        const segments = []
        for (let index = trimLeft; index <= contentArray.length; ) {
            const nextPointOfInterest = highlightRanges.shift()

            // There are no more highlight segments, so we can render the remaining text
            if (nextPointOfInterest === undefined) {
                segments.push(contentArray.slice(index).join(''))
                break
            } else {
                // There are highlight segments, so we need to render the text before the next
                // highlight and then render the highlight.
                const [offset, length] = nextPointOfInterest
                const rest = contentArray.slice(index, offset).join('')
                if (rest !== '') {
                    segments.push(rest)
                }

                // Note: The <span> might cut-off a joined emoji sequence in-between a &zwj;. For
                // example if the content contains "👨‍👩‍👧‍👧" but we search for "👨‍👩‍👧", we will highlight
                // only a section of the joined sequence, e.g.: "<span>👨‍👩‍👧</span>&zwj;‍👧".
                segments.push(
                    <span key={`highlight-${index}`} className={styles.codeHighlight}>
                        {contentArray.slice(offset, offset + length).join('')}
                    </span>
                )
                index = offset + length
            }
        }

        return <div className="text-truncate">{segments}</div>
    }
)
