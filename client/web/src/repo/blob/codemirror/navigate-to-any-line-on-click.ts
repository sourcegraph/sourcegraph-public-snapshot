import { Extension, RangeSetBuilder } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, PluginValue, ViewPlugin, ViewUpdate } from '@codemirror/view'
import * as H from 'history'

import { addLineRangeQueryParameter, formatSearchParameters } from '@sourcegraph/common'

// Extension that is only used in the ref panel preview pane. Wraps all lines
// with an anchor link so that clicking on the line promotes that line from the
// preview pane into the main blob view.
export function navigateToLineOnAnyClickExtension(
    location: H.Location,
    history: H.History,
    pushURL: (url: string) => void
): Extension {
    class LineLinkManager implements PluginValue {
        public decorations: DecorationSet = Decoration.none
        constructor(view: EditorView) {
            this.decorations = this.computeDecorations(view)
        }
        public update(update: ViewUpdate): void {
            if (update.viewportChanged) {
                this.decorations = this.computeDecorations(update.view)
            }
        }

        private computeDecorations(view: EditorView): DecorationSet {
            const builder = new RangeSetBuilder<Decoration>()
            const startLine = view.state.doc.lineAt(view.viewport.from)
            const endLine = view.state.doc.lineAt(view.viewport.to)
            const parameters = new URLSearchParams(location.search)
            parameters.delete('popover')
            for (let lineNumber = startLine.number; lineNumber <= endLine.number; lineNumber++) {
                const line = view.state.doc.line(lineNumber)
                const entry: H.LocationDescriptor<unknown> = {
                    ...location,
                    search: formatSearchParameters(addLineRangeQueryParameter(parameters, `L${line.number}`)),
                }
                const url = history.createHref(entry)
                builder.add(
                    line.from,
                    line.to,
                    Decoration.mark({ tagName: 'a', attributes: { href: url, 'data-line-link': '' } })
                )
            }
            return builder.finish()
        }
    }
    return [
        ViewPlugin.fromClass(LineLinkManager, { decorations: plugin => plugin.decorations }),
        EditorView.domEventHandlers({
            click(event: MouseEvent) {
                const target = event.target as HTMLElement
                // Check to see if the clicked target is a token link.
                // If it is, push the link to the history stack.
                if (target.matches('[data-line-link]')) {
                    event.preventDefault()
                    pushURL(target.getAttribute('href')!)
                }
            },
        }),
    ]
}
