/**
 * This provides CodeMirror extension for highlighting a static set of ranges.
 */
import { type Extension } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'

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

export function staticHighlights(ranges: Range[]): Extension {
    return [
        EditorView.decorations.compute(['doc'], state =>
            Decoration.set(
                ranges.map(range =>
                    staticHighlightDecoration.range(
                        // Codemirror expects 1-based line numbers
                        state.doc.line(range.start.line + 1).from + range.start.column,
                        state.doc.line(range.end.line + 1).from + range.end.column
                    )
                ),
                true
            )
        ),
        EditorView.theme({
            '.cm-sg-staticSelection': {
                backgroundColor: 'var(--mark-bg)',
            },
        }),
    ]
}
