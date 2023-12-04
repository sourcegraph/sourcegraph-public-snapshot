import { type Extension, RangeSetBuilder } from '@codemirror/state'
import {
    Decoration,
    type DecorationSet,
    EditorView,
    type PluginValue,
    ViewPlugin,
    type ViewUpdate,
} from '@codemirror/view'
import type { NavigateFunction } from 'react-router-dom'

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
                        attributes: { href, 'data-cm-line-link': '' },
                        class: 'text-decoration-none text-inherit',
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
export function navigateToLineOnAnyClickExtension(navigate: NavigateFunction): Extension {
    return [
        ViewPlugin.fromClass(LineLinkManager, { decorations: plugin => plugin.decorations }),
        EditorView.domEventHandlers({
            click(event) {
                const target = event.target as HTMLElement
                const closest = target.closest('[data-cm-line-link]')

                // Check to see if the clicked target is a or is inside a token link.
                // If it is, navigate to the link.
                if (closest) {
                    event.preventDefault()
                    navigate(closest.getAttribute('href')!)
                }
            },
        }),
    ]
}
