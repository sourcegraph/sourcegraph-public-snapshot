import { Annotation, Extension, RangeSet, Range, RangeSetBuilder, StateEffect, StateField } from '@codemirror/state'
import {
    EditorView,
    Decoration,
    lineNumbers,
    ViewPlugin,
    PluginValue,
    ViewUpdate,
    GutterMarker,
    gutterLineClass,
} from '@codemirror/view'

/**
 * Represents the currently selected line range. null means no lines are
 * selected. Line numbers are 1-based.
 * endLine may be smaller than line
 */
export type SelectedLineRange = { line: number; endLine?: number } | null

const selectedLineDecoration = Decoration.line({ class: 'selected-line' })
const selectedLineGutterMarker = new (class extends GutterMarker {
    public elementClass = 'selected-line'
})()
const setSelectedLines = StateEffect.define<SelectedLineRange>()
const setEndLine = StateEffect.define<number>()

/**
 * This field stores the selected line range and provides the corresponding line
 * decorations.
 */
export const selectedLines = StateField.define<SelectedLineRange>({
    create() {
        return null
    },
    compare(previous, next): boolean {
        return previous?.line === next?.line && previous?.endLine === next?.endLine
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

            for (let lineNumber = from; lineNumber <= to; lineNumber++) {
                const from = state.doc.line(lineNumber).from
                builder.add(from, from, selectedLineDecoration)
            }

            return builder.finish()
        }),
        gutterLineClass.compute([field], state => {
            const range = state.field(field)
            const marks: Range<GutterMarker>[] = []

            if (range) {
                const endLine = range.endLine ?? range.line
                const from = Math.min(range.line, endLine)
                const to = from === endLine ? range.line : endLine

                for (let lineNumber = from; lineNumber <= to; lineNumber++) {
                    marks.push(selectedLineGutterMarker.range(state.doc.line(lineNumber).from))
                }
            }

            return RangeSet.of(marks)
        }),
    ],
})

/**
 * An annotation to indicate where a line selection is comming from.
 * Transactions that set selected lines without this annotion are assumed to be
 * "external" (e.g. from syncing with the URL).
 */
const lineSelectionSource = Annotation.define<'gutter'>()

/**
 * View plugin resonsible for scrolling the selected line(s) into view if/when
 * necessary.
 */
const scrollIntoView = ViewPlugin.fromClass(
    class implements PluginValue {
        private lastSelectedLines: SelectedLineRange | null = null
        constructor(private readonly view: EditorView) {}

        public update(update: ViewUpdate): void {
            const currentSelectedLines = update.state.field(selectedLines)
            if (
                this.lastSelectedLines !== currentSelectedLines &&
                update.transactions.some(transaction => transaction.annotation(lineSelectionSource) !== 'gutter')
            ) {
                // Only scroll selected lines into view when the user isn't
                // currently selecting lines themselves (as indicated by the
                // presence of the "gutter" annotation). Otherwise the scroll
                // position might change while the user is selecting lines.
                this.lastSelectedLines = currentSelectedLines
                this.scrollIntoView(currentSelectedLines)
            }
        }

        public scrollIntoView(selection: SelectedLineRange): void {
            if (selection && shouldScrollIntoView(this.view, selection)) {
                window.requestAnimationFrame(() => {
                    this.view.dispatch({
                        effects: EditorView.scrollIntoView(this.view.state.doc.line(selection.line).from, {
                            y: 'center',
                        }),
                    })
                })
            }
        }
    }
)

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
export function selectableLineNumbers(config: {
    onSelection: (range: SelectedLineRange) => void
    initialSelection: SelectedLineRange | null
}): Extension {
    let dragging = false

    return [
        scrollIntoView,
        selectedLines.init(() => config.initialSelection),
        lineNumbers({
            domEventHandlers: {
                mousedown(view, block, event) {
                    const line = view.state.doc.lineAt(block.from).number
                    const range = view.state.field(selectedLines)

                    view.dispatch({
                        effects: (event as MouseEvent).shiftKey
                            ? setEndLine.of(line)
                            : setSelectedLines.of(isSingleLine(range) && range?.line === line ? null : { line }),
                        annotations: lineSelectionSource.of('gutter'),
                    })

                    dragging = true

                    function onmouseup(): void {
                        dragging = false
                        window.removeEventListener('mouseup', onmouseup)
                        window.removeEventListener('mousemove', onmousemove)

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

                    function onmousemove(event: MouseEvent): void {
                        if (dragging) {
                            const newEndline = view.state.doc.lineAt(view.posAtCoords(event, false)).number
                            if (view.state.field(selectedLines)?.endLine !== newEndline) {
                                view.dispatch({
                                    effects: setEndLine.of(newEndline),
                                    annotations: lineSelectionSource.of('gutter'),
                                })
                            }
                            event.preventDefault()
                        }
                    }

                    window.addEventListener('mouseup', onmouseup)
                    window.addEventListener('mousemove', onmousemove)
                    return true
                },
            },
        }),
        EditorView.theme({
            '.cm-lineNumbers': {
                cursor: 'pointer',
                color: 'var(--line-number-color)',
            },
            '.cm-lineNumbers .cm-gutterElement:hover': {
                textDecoration: 'underline',
            },
        }),
    ]
}

/**
 * Set selected lines (e.g. from the URL).
 */
export function selectLines(view: EditorView, newRange: SelectedLineRange): void {
    view.dispatch({ effects: setSelectedLines.of(newRange) })
}

/**
 * This function determines whether or not the selected lines are in view by
 * comparing the top/bottom positions of the line (which are relative to the
 * document top) to the scroll position of the scroll container.
 *
 * Simply using EditorView.viewport doesn't work because those returns the range
 * of *rendered* lines, not just *visible* lines (some lines are rendered
 * outside of the editor viewport).
 */
function shouldScrollIntoView(view: EditorView, range: SelectedLineRange): boolean {
    if (!range) {
        return false
    }

    const from = view.lineBlockAt(view.state.doc.line(range.line).from)
    const to = range.endLine ? view.lineBlockAt(view.state.doc.line(range.endLine).to) : from

    return (
        from.top + from.height >= view.scrollDOM.scrollTop + view.scrollDOM.clientHeight ||
        to.top <= view.scrollDOM.scrollTop
    )
}

function isSingleLine(range: SelectedLineRange): boolean {
    return !!range && (!range.endLine || range.line === range.endLine)
}
