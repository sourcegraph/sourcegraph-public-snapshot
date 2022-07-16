import { Extension, RangeSetBuilder, StateEffect, StateField } from '@codemirror/state'
import { EditorView, Decoration, lineNumbers } from '@codemirror/view'

/**
 * Represents the currently selected line range. null means no lines are
 * selected. Line numbers are 1-based.
 * endLine may be smaller than line
 */
export type SelectedLineRange = { line: number; endLine?: number } | null

const highlighedLineDecoration = Decoration.line({ class: 'selected-line' })
const setSelectedLines = StateEffect.define<SelectedLineRange>()
const setEndLine = StateEffect.define<number>()

/**
 * This field stores the selected line range and provides the corresponding line
 * decorations.
 */
const selectedLines = StateField.define<SelectedLineRange>({
    create() {
        return null
    },
    update(value, transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(setSelectedLines)) {
                return effect.value
            }
            if (effect.is(setEndLine)) {
                if (!value?.line) {
                    value = { line: effect.value }
                }
                return { ...value, endLine: effect.value }
            }
        }
        return value
    },
    provide: field => [
        EditorView.decorations.compute([field], state => {
            const range = state.field(field)
            if (!range) {
                return Decoration.none
            }

            // By ordering line and endLine here we make "inverse" selection
            // work automagically

            const endLine = range.endLine ?? range.line
            const from = Math.min(range.line, endLine)
            const to = from === endLine ? range.line : endLine

            const builder = new RangeSetBuilder<Decoration>()

            for (let line = from; line <= to; line++) {
                const from = state.doc.line(line).from
                builder.add(from, from, highlighedLineDecoration)
            }

            return builder.finish()
        }),
    ],
})

/**
 * This extension provides a line gutter that allows selecting (ranges of) lines
 * by clicking (and dragging over) the line numbers. Shift+click to select a
 * range is also supported.
 *
 * onSelection is called when a selection was made. range.line will always be <
 * range.endLine.
 *
 * NOTE: Dragging to select on the gutter won't automatically scroll the
 * document.
 */
export function selectableLineNumbers(config: { onSelection: (range: SelectedLineRange) => void }): Extension {
    let dragging = false

    return [
        lineNumbers({
            domEventHandlers: {
                mousedown(view, block, event) {
                    const line = view.state.doc.lineAt(block.from).number
                    const range = view.state.field(selectedLines)

                    view.dispatch({
                        effects: (event as MouseEvent).shiftKey
                            ? setEndLine.of(line)
                            : setSelectedLines.of(isSingleLine(range) && range?.line === line ? null : { line }),
                    })

                    dragging = true

                    function onmouseup(): void {
                        dragging = false
                        window.removeEventListener('mouseup', onmouseup)

                        let range = view.state.field(selectedLines)
                        if (range) {
                            // Order line and endLine
                            if (range.endLine && range.line > range.endLine) {
                                range = {
                                    line: range.endLine,
                                    endLine: range.line,
                                }
                            } else if (range.line === range.endLine) {
                                range = { line: range.line }
                            } else {
                                range = { ...range }
                            }
                        }
                        config.onSelection(range)
                    }
                    window.addEventListener('mouseup', onmouseup)
                    return true
                },
                mousemove(view, line) {
                    if (dragging) {
                        const newEndline = view.state.doc.lineAt(line.from).number
                        const { endLine } = view.state.field(selectedLines) ?? {}
                        if (endLine !== newEndline) {
                            view.dispatch({ effects: setEndLine.of(newEndline) })
                        }
                        return true
                    }
                    return false
                },
            },
        }),
        selectedLines,
        EditorView.theme({
            '.cm-lineNumbers': {
                cursor: 'pointer',
            },
            '.cm-lineNumbers .cm-gutterElement:hover': {
                textDecoration: 'underline',
            },
        }),
    ]
}

/**
 * Set selected lines (e.g. from the URL). The function won't trigger an update
 * if the same lines are already selected.
 */
export function selectLines(view: EditorView, newRange: SelectedLineRange): void {
    const currentRange = view.state.field(selectedLines)

    if (currentRange?.line === newRange?.line && currentRange?.endLine === newRange?.endLine) {
        return
    }

    const effects: StateEffect<unknown>[] = [setSelectedLines.of(newRange)]

    if (newRange) {
        effects.push(
            EditorView.scrollIntoView(
                view.state.doc.line(newRange.line).from,
                // This is not ideal but  shouldScrollIntoView operates on the
                // rendered lines, of which not all might be in view. In that case
                // "nearest" ensures that the line will be visible but it won't have
                // any affect if the line is already visible.
                { y: shouldScrollIntoView(view, newRange) ? 'center' : 'nearest' }
            )
        )
    }

    view.dispatch({ effects })
}

function shouldScrollIntoView(view: EditorView, range: SelectedLineRange): boolean {
    if (!range) {
        return false
    }

    const from = view.state.doc.line(range.line).from
    const to = range.endLine ? view.state.doc.line(range.endLine).to : undefined

    return from >= view.viewport.to || (to ?? from) <= view.viewport.from
}

function isSingleLine(range: SelectedLineRange): boolean {
    return !!range && (!range.endLine || range.line === range.endLine)
}
