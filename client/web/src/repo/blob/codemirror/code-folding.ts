import { EditorState, Extension, Line, StateField } from '@codemirror/state'
import { foldGutter, foldService } from '@codemirror/language'

enum CharCode {
    /**
     * The `\t` character.
     */
    Tab = 9,

    Space = 32,
}

/**
 * Returns:
 *  - -1 => the line consists of whitespace
 *  - otherwise => the indent level is returned value
 */
function computeIndentLevel(line: string, tabSize: number): number {
    let indent = 0
    let i = 0
    const len = line.length

    while (i < len) {
        const charCode = line.charCodeAt(i)

        if (charCode === CharCode.Space) {
            indent++
        } else if (charCode === CharCode.Tab) {
            indent = indent - (indent % tabSize) + tabSize
        } else {
            break
        }
        i++
    }

    if (i === len) {
        return -1 // line only consists of whitespace
    }

    return indent
}

/**
 * Stores folding ranges
 * Computes folding ranges  Stores folding ranges  focused (selected), hovered and pinned {@link Occurrence}s and {@link Tooltip}s associate with them.
 */
const foldingRanges = StateField.define<[Line, Line][]>({
    create(state: EditorState) {
        const ranges: [Line, Line][] = []
        const previousRanges = [{ indent: -1, endAbove: state.doc.lines + 1 }]

        for (let lineNumber = state.doc.lines; lineNumber > 0; lineNumber--) {
            const line = state.doc.line(lineNumber)
            const indent = computeIndentLevel(line.text, state.tabSize)
            if (indent === -1) {
                continue
            }

            let previous = previousRanges[previousRanges.length - 1]
            if (previous.indent > indent) {
                // remove ranges with larger indent
                do {
                    previousRanges.pop()
                    previous = previousRanges[previousRanges.length - 1]
                } while (previous.indent > indent)

                // new folding range
                const endLineNumber = previous.endAbove - 1
                if (endLineNumber - lineNumber >= 1) {
                    // should be at least 2 lines
                    ranges.push([line, state.doc.line(endLineNumber)])
                }
            }
            if (previous.indent === indent) {
                previous.endAbove = lineNumber
            } else {
                // previous.indent < indent
                // new range with a bigger indent
                previousRanges.push({ indent, endAbove: lineNumber })
            }
        }

        return ranges
    },
    update(value, transaction) {
        return value
    },
})

function getFoldRange(state: EditorState, lineStart: number): { from: number; to: number } | null {
    const ranges = state.field(foldingRanges)

    const range = ranges.find(([start]) => {
        return start.number === state.doc.lineAt(lineStart).number
    })

    if (!range) {
        return null
    }

    const [start, end] = range

    return { from: start.to, to: end.to }
}

export function codeFoldingExtension(): Extension {
    return [foldingRanges, foldGutter(), foldService.of(getFoldRange)]
}
