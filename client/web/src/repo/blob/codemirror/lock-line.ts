import type { Line, StateEffect } from '@codemirror/state'
import { EditorView } from '@codemirror/view'

import { shouldScrollIntoView } from './linenumbers'

// Avoid jumpy scrollbar when the line dimensions change by locking on to the
// first visible line.
//
// Details https://github.com/sourcegraph/sourcegraph/issues/41413
export function lockFirstVisibleLine(view: EditorView): StateEffect<unknown>[] {
    const firstLine = firstVisibleLine(view)
    if (firstLine) {
        return [EditorView.scrollIntoView(firstLine.from, { y: 'start' })]
    }
    return []
}

// Returns the first line that is visible in the editor. We can't directly use
// the viewport for this functionality because the viewport includes lines that
// are rendered but not visible.
function firstVisibleLine(view: EditorView): Line | undefined {
    for (const { from, to } of view.visibleRanges) {
        for (let pos = from; pos < to; ) {
            const line = view.state.doc.lineAt(pos)
            // This may be an inefficient way to detect the first visible line
            // but it appears to work correctly and this is unlikely to be a
            // performance bottleneck since we should only use need to compute
            // this for infrequently used code-paths like when enabling/disabling
            // line wrapping, or when lazy syntax highlighting gets loaded.
            if (!shouldScrollIntoView(view, { line: line.number + 1 })) {
                return line
            }
            pos = line.to + 1
        }
    }
    return undefined
}
