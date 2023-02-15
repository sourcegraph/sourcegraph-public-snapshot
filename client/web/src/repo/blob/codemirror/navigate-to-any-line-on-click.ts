import { Extension, RangeSetBuilder } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, PluginValue, ViewPlugin, ViewUpdate } from '@codemirror/view'

import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { blobPropsFacet } from '.'

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
        const blobInfo = view.state.facet(blobPropsFacet).blobInfo
        for (const { from, to } of view.visibleRanges) {
            for (let pos = from; pos < to; ) {
                const line = view.state.doc.lineAt(pos)
                const href = toPrettyBlobURL({
                    ...blobInfo,
                    position: { line: line.number, character: 0 },
                })
                builder.add(
                    line.from,
                    line.to,
                    Decoration.mark({
                        tagName: 'a',
                        attributes: { href, 'data-line-link': '' },
                        class: 'text-decoration-none',
                    })
                )
                pos = line.to + 1
            }
        }
        return builder.finish()
    }
}

// Extension that is only used in the ref panel preview pane. Wraps all lines
// with an anchor link so that clicking on the line promotes that line from the
// preview pane into the main blob view.
export const navigateToLineOnAnyClickExtension: Extension = [
    ViewPlugin.fromClass(LineLinkManager, { decorations: plugin => plugin.decorations }),
    EditorView.domEventHandlers({
        click(event, view) {
            const target = event.target as HTMLElement
            // Check to see if the clicked target is a token link.
            // If it is, push the link to the history stack.
            if (target.matches('[data-line-link]')) {
                event.preventDefault()
                const href = target.getAttribute('href')!
                const props = view.state.facet(blobPropsFacet)
                if (props.nav) {
                    props.nav(href)
                } else {
                    props.navigate(href)
                }
            }
        },
    }),
]
