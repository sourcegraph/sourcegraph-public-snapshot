import { Extension, Facet, StateField } from '@codemirror/state'
import { LineOrPositionOrRange } from '@sourcegraph/common'
import { UIPositionSpec } from '@sourcegraph/shared/src/util/url'
import { BlobProps } from '../Blob'
import { hovercardRanges } from './hovercard'
import { offsetToPosition, uiPositionToOffset } from './utils'

/**
 * This facet is a necessary evil to allow access to props passed to the blob
 * component from React components rendered by extensions.
 */
export const blobPropsFacet = Facet.define<BlobProps, BlobProps>({
    combine: props => props[0],
})

export function hovercardRangeFromPin(field: StateField<LineOrPositionOrRange | null>): Extension {
    return hovercardRanges.computeN([field], state => {
        const position = state.field(field)
        if (!position) {
            return []
        }

        if (!position.line || !position.character) {
            return []
        }
        const startLine = state.doc.line(position.line)

        const startPosition = {
            line: position.line,
            character: position.character,
        }
        const from = uiPositionToOffset(state.doc, startPosition, startLine)

        let endPosition: UIPositionSpec['position']
        let to: number

        if (position.endLine && position.endCharacter) {
            endPosition = {
                line: position.endLine,
                character: position.endCharacter,
            }
            to = uiPositionToOffset(state.doc, endPosition)
        } else {
            // To determine the end position we have to find the word at the
            // start position
            const word = state.wordAt(from)
            if (!word) {
                return []
            }
            to = word.to
            endPosition = offsetToPosition(state.doc, word.to)
        }

        return [
            {
                to,
                from,
                range: {
                    start: startPosition,
                    end: endPosition,
                },
            },
        ]
    })
}
