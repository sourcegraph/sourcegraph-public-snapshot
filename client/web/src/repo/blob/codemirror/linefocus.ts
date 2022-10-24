import { Extension, Line, RangeSetBuilder } from '@codemirror/state'
import { EditorView, Decoration, PluginValue, ViewUpdate, ViewPlugin } from '@codemirror/view'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { fromEvent, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'

import { SelectedLineRange, selectedLines, setSelectedLines } from './linenumbers'

interface FocusableLinesConfig {
    initialLine?: number
    onSelection: (range: SelectedLineRange) => void
}

const focusableLineDecoration = Decoration.line({ attributes: { tabIndex: '-1' } })

export const [focusedLine, updateFocusedLine] = createUpdateableField<number | null>(null)

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
        // focusedLine.init(() => initialLine ?? null),
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

                    focusedLineNumber = nextLineNumber
                    view.requestMeasure({
                        read(view) {
                            return view.domAtPos(nextLine.from).node as HTMLElement | null
                        },
                        write(measure) {
                            if (measure) {
                                measure.focus()
                            }
                        },
                    })
                }

                if (event.key === 'Enter') {
                    const isLink = event.target instanceof HTMLAnchorElement
                    const currentLine = view.state.field(focusedLine)
                    if (!isLink && currentLine) {
                        view.dispatch({
                            effects: setSelectedLines.of({ line: currentLine }),
                        })
                        onSelection({ line: currentLine })
                    }
                }
            },
        }),
        // ViewPlugin.define(view => new LineFocusManager(view)),
    ]
}
