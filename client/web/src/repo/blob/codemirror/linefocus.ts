import { Extension, RangeSetBuilder } from '@codemirror/state'
import { EditorView, Decoration } from '@codemirror/view'

import { SelectedLineRange, setSelectedLines } from './linenumbers'

interface FocusableLinesConfig {
    initialLine?: number
    onSelection: (range: SelectedLineRange) => void
}

const focusableLineDecoration = Decoration.line({ attributes: { tabIndex: '-1' } })

export function focusableLines({ initialLine, onSelection }: FocusableLinesConfig): Extension {
    let focusedLineNumber: number = initialLine ?? 1

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
            keydown(event: KeyboardEvent, view: EditorView) {
                if (event.key === 'ArrowUp' || event.key === 'ArrowDown') {
                    event.preventDefault()

                    const prevLine = focusedLineNumber
                    const nextLineNumber =
                        event.key === 'ArrowUp'
                            ? Math.max(1, prevLine - 1)
                            : Math.min(prevLine + 1, view.state.doc.lines)

                    const nextLine = view.state.doc.line(nextLineNumber)
                    const nextLineElement = view.domAtPos(nextLine.from).node as HTMLElement | null

                    if (nextLineElement) {
                        focusedLineNumber = nextLineNumber
                        window.requestAnimationFrame(() => {
                            nextLineElement.focus()
                        })
                    }
                }

                if (event.key === 'Enter') {
                    const isLink = event.target instanceof HTMLAnchorElement
                    if (!isLink && focusedLineNumber) {
                        view.dispatch({
                            effects: setSelectedLines.of({ line: focusedLineNumber }),
                        })
                        onSelection({ line: focusedLineNumber })
                    }
                }
            },
            focusin(event: FocusEvent, view: EditorView) {
                const currentFocus = event.target as HTMLElement | null

                if (currentFocus) {
                    const nearestLine = view.state.doc.lineAt(view.posAtDOM(currentFocus))
                    focusedLineNumber = nearestLine.number
                }
            },
        }),
    ]
}
