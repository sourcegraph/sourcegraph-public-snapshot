import {
    Annotation,
    EditorState,
    Extension,
    Range,
    RangeSet,
    RangeSetBuilder,
    StateEffect,
    StateField,
} from '@codemirror/state'
import {
    Decoration,
    EditorView,
    gutterLineClass,
    GutterMarker,
    lineNumbers,
    PluginValue,
    ViewPlugin,
    ViewUpdate,
} from '@codemirror/view'

import { isValidLineRange, MOUSE_MAIN_BUTTON, preciseOffsetAtCoords } from './utils'

/**
 * Represents the currently selected line range. null means no lines are
 * selected. Line numbers are 1-based.
 * endLine may be smaller than line
 */
export type SelectedLineRange = { line: number; character?: number; endLine?: number } | null

const selectedLineDecoration = Decoration.line({
    class: 'selected-line',
    attributes: { tabIndex: '-1', 'data-line-focusable': '' },
})
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
            const to = Math.min(state.doc.lines, from === endLine ? range.line : endLine)

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
                const to = Math.min(state.doc.lines, from === endLine ? range.line : endLine)

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
        constructor(private readonly view: EditorView) {
            this.lastSelectedLines = this.view.state.field(selectedLines)
            this.scrollIntoView(this.lastSelectedLines)
        }

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
 * This plugin handles selecting lines by clicking on the end empty after them.
 * What makes this complex is handling text selection properly, and not such
 * figuring out that a user is selecting text (that's easy) but to prevent text
 * selection from being rendered if the user actually want to select multiple
 * lines by shift clicking.
 *
 * Desired behavior:
 * - Drag to select text
 * - Click to select line
 * - Shift click to select text when there is already other selected text
 * - Shift click to select line range if there is no selected text
 */
function selectOnClick({ onSelection }: SelectableLineNumbersConfig): Extension {
    // Maybe it would be better to use state fields for this (I don't know). It
    // works though.
    let maybeSelectLine = false
    let preventTextSelection = false

    return [
        EditorState.transactionFilter.of(transaction => {
            // If the user tries to select a text range (and doesn't just click
            // somewhere)
            if (
                transaction.isUserEvent('select') &&
                transaction.selection &&
                transaction.selection.main.from !== transaction.selection.main.to
            ) {
                if (preventTextSelection) {
                    return []
                }
                // If we are selecting a text range and not already prevent text
                // selection then we don't want to select a line.
                maybeSelectLine = false
            }
            return transaction
        }),
        EditorView.domEventHandlers({
            mousedown(event, view) {
                if (event.button !== MOUSE_MAIN_BUTTON) {
                    // Only handle clicks with the main button
                    return
                }

                maybeSelectLine = true
                preventTextSelection = false

                if (event.shiftKey) {
                    // Selecting text via shift click is only supported when
                    // there is already other selected text.
                    if (hasTextSelection(view.state)) {
                        maybeSelectLine = false
                    } else {
                        // Otherwise we need to prevent CodeMirror/the browser
                        // from applying text selection
                        preventTextSelection = true
                    }
                }
            },
            mouseup(event, view) {
                preventTextSelection = false

                if (!maybeSelectLine || event.button !== MOUSE_MAIN_BUTTON) {
                    return
                }

                maybeSelectLine = false

                // IMPORTANT: This gives the offset of the character *closest*
                // to the clicked position, not *at* the clicked position.
                const offset = view.posAtCoords(event)
                // Ignore clicks outside the document
                if (offset === null) {
                    return
                }

                let selectedLine: number | null = null

                const clickedLine = view.state.doc.lineAt(offset)
                if (offset === clickedLine.to) {
                    // If the offset is the same value as the end position of
                    // the line then click happend after the last charcater.
                    selectedLine = clickedLine.number
                } else if (offset === clickedLine.from && preciseOffsetAtCoords(view, event) === null) {
                    // `preciseOffsetAtCoords(...) === null` allows us to recognize clicks before the actual text content
                    // while `offset === clickedLine.from` ensures that we ignore clicks between lines
                    selectedLine = clickedLine.number
                }

                if (selectedLine !== null) {
                    view.dispatch({
                        effects: event.shiftKey
                            ? setEndLine.of(selectedLine)
                            : setSelectedLines.of({ line: selectedLine }),
                    })
                    onSelection(normalizeLineRange(view.state.field(selectedLines)))
                }
            },
        }),
    ]
}

interface SelectableLineNumbersConfig {
    onSelection: (range: SelectedLineRange) => void
    initialSelection: SelectedLineRange | null
    navigateToLineOnAnyClick: boolean
    enableSelectionDrivenCodeNavigation?: boolean
}

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
export function selectableLineNumbers(config: SelectableLineNumbersConfig): Extension {
    let dragging = false

    return [
        scrollIntoView,
        selectedLines.init(() => config.initialSelection),
        lineNumbers({
            domEventHandlers: {
                mousedown(view, block, event) {
                    const mouseEvent = event as MouseEvent // make TypeScript happy
                    if (mouseEvent.button !== MOUSE_MAIN_BUTTON) {
                        return false
                    }

                    const line = view.state.doc.lineAt(block.from).number
                    if (config.navigateToLineOnAnyClick) {
                        // Only support single line selection when navigateToLineOnAnyClick is true.
                        view.dispatch({
                            effects: setSelectedLines.of({ line }),
                            annotations: lineSelectionSource.of('gutter'),
                        })
                    } else {
                        const range = view.state.field(selectedLines)
                        view.dispatch({
                            effects: mouseEvent.shiftKey
                                ? setEndLine.of(line)
                                : setSelectedLines.of(isSingleLine(range) && range?.line === line ? null : { line }),
                            annotations: lineSelectionSource.of('gutter'),
                            // Collapse/reset text selection
                            selection: { anchor: view.state.selection.main.anchor },
                        })
                    }

                    dragging = true

                    function onmouseup(event: MouseEvent): void {
                        if (event.button !== MOUSE_MAIN_BUTTON) {
                            return
                        }

                        dragging = false
                        window.removeEventListener('mouseup', onmouseup)
                        window.removeEventListener('mousemove', onmousemove)
                        config.onSelection(normalizeLineRange(view.state.field(selectedLines)))
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
                            // Prevents the browser from selecting the line
                            // numbers as text
                            event.preventDefault()
                        }
                    }

                    window.addEventListener('mouseup', onmouseup)
                    window.addEventListener('mousemove', onmousemove)
                    return true
                },
            },
        }),
        // Disable `selectOnClick` with token selection because they interact
        // badly with each other causing errors.
        config.enableSelectionDrivenCodeNavigation ? [] : selectOnClick(config),
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
    view.dispatch({
        effects: setSelectedLines.of(newRange && isValidLineRange(newRange, view.state.doc) ? newRange : null),
    })
}

function normalizeLineRange(range: SelectedLineRange): SelectedLineRange {
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
    return range
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
export function shouldScrollIntoView(view: EditorView, range: SelectedLineRange): boolean {
    if (!range || !isValidLineRange(range, view.state.doc)) {
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

/**
 * Helper function that returns true if the user has selected any text in the
 * document. A CodeMirror always has a "selection", which determines the cursor
 * position but only if its start and end are different it actually represents
 * selected text.
 */
function hasTextSelection(state: EditorState): boolean {
    const range = state.selection.asSingle().main
    return range.from !== range.to
}
