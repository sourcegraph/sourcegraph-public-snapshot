import { Extension, Facet, Line, RangeSetBuilder } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'

import { SelectedLineRange, setSelectedLines } from './linenumbers'

const focusableLineDecoration = Decoration.line({ class: 'sourcegraph-line-focus', attributes: { tabIndex: '-1' } })

function arrowKeyNav(): Extension {
    let activeLine: Line

    return [
        EditorView.decorations.compute([], state => {
            const to = state.doc.lines
            const builder = new RangeSetBuilder<Decoration>()

            for (let lineNumber = 1; lineNumber <= to; lineNumber++) {
                const from = state.doc.line(lineNumber).from
                builder.add(from, from, focusableLineDecoration)
            }

            return builder.finish()
        }),
        EditorView.domEventHandlers({
            focusin: (event: FocusEvent, view: EditorView) => {
                const currentFocus = event.target as HTMLElement
                const position = view.posAtDOM(currentFocus)
                activeLine = view.state.doc.lineAt(position)
            },
            keydown: (event: KeyboardEvent, view: EditorView) => {
                if (event.key === 'ArrowUp' || event.key === 'ArrowDown') {
                    event.preventDefault()
                    if (activeLine) {
                        const nextLineNumber = event.key === 'ArrowUp' ? activeLine.number - 1 : activeLine.number + 1
                        const nextLine = view.state.doc.line(nextLineNumber)
                        const nextLineElement = view.domAtPos(nextLine.from).node as HTMLElement

                        if (nextLineElement) {
                            activeLine = nextLine
                            window.requestAnimationFrame(() => {
                                nextLineElement.focus()
                            })
                        }
                    }
                }
            },
            keyup: (event: KeyboardEvent, view: EditorView) => {
                // TODO: Do this logic on keydown, fix issue with conflicting event handlers on links
                if (event.key === 'Enter') {
                    if (activeLine) {
                        view.dispatch({
                            effects: setSelectedLines.of({ line: activeLine.number }),
                        })
                        const { onSelection } = view.state.facet(arrowKeyNavigation)
                        onSelection({ line: activeLine.number })
                    }
                }
            },
        }),
    ]
}

interface ArrowKeyNavigationFacet {
    onSelection: (range: SelectedLineRange) => void
}

/**
 * Facet with which we can provide `BlobInfo`, specifically `stencil` ranges.
 *
 * This enables the `tokenLinks` extension which will decorate tokens with links
 * to either their definition or the references panel.
 */
export const arrowKeyNavigation = Facet.define<ArrowKeyNavigationFacet, ArrowKeyNavigationFacet>({
    combine: source => source[0],
    enables: arrowKeyNav(),
})
