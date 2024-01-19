/**
 * This provides CodeMirror extension for highlighting a static set of ranges.
 */
import { EditorState, type Extension } from '@codemirror/state'
import { Decoration, EditorView, EditorViewConfig } from '@codemirror/view'
import { sortBy } from 'lodash'

export interface Range {
    start: Location
    end: Location
}

export interface Location {
    // A zero-based line number
    line: number
    // A zero-based column number
    column: number
}

const staticHighlightDecoration = Decoration.mark({ class: 'cm-sg-staticSelection' })

export function staticHighlights(ranges: Range[], scrollIntoView?: boolean): Extension {
    sortBy(ranges, [range => [range.start.line, range.start.column, range.end.line, range.end.column]])

    return [
        EditorView.decorations.compute(['doc'], state =>
            Decoration.set(
                ranges.map(range =>
                    staticHighlightDecoration.range(
                        // Codemirror expects 1-based line numbers
                        state.doc.line(range.start.line + 1).from + range.start.column,
                        state.doc.line(range.end.line + 1).from + range.end.column
                    )
                )
            )
        ),
        EditorView.theme({
            '.cm-sg-staticSelection': {
                backgroundColor: 'var(--mark-bg)',
            },
        }),
    ]
}
