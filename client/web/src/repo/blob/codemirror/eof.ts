import type { Extension } from '@codemirror/state'
import { Decoration, EditorView, WidgetType } from '@codemirror/view'

class NoLineBreakWidget extends WidgetType {
    constructor(private noLineBreakComment: string) {
        super()
    }

    public eq(other: NoLineBreakWidget): boolean {
        return this.noLineBreakComment === other.noLineBreakComment
    }

    public toDOM(): HTMLElement {
        const div = document.createElement('div')
        div.className = 'no-line-break-msg'
        div.textContent = this.noLineBreakComment
        return div
    }
}

// This decoration will replace/remove the last line break which
// causes the last line to be hidden.
const removeLastLineDeco = Decoration.replace({})
const eofNoteDeco = Decoration.replace({
    widget: new NoLineBreakWidget('(No newline at end of file)'),
    block: true,
})

/**
 * An extensions that hides the last (empty) line if the file ends with
 * a line break, or shows a message that the line doesn't end with a line
 * break.
 */
export const hideEmptyLastLine: Extension = [
    EditorView.decorations.compute(['doc'], state => {
        const lastLine = state.doc.line(state.doc.lines)
        return Decoration.set(
            lastLine.length === 0
                ? // Subtract 1 to exclude newline character at end of line
                  // when setting decoration range
                  removeLastLineDeco.range(lastLine.from - 1, lastLine.to)
                : eofNoteDeco.range(lastLine.to)
        )
    }),
    EditorView.theme({
        '.no-line-break-msg': {
            color: 'var(--text-muted)',
            fontStyle: 'italic',
            marginTop: '.2rem',
            userSelect: 'none',
            '-webkit-user-select': 'none',
        },
    }),
]
