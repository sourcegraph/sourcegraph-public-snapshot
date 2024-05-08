import {
    Annotation,
    type Extension,
    type Range,
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
    layer,
    lineNumbers,
    type PluginValue,
    RectangleMarker,
    ViewPlugin,
    type ViewUpdate,
} from '@codemirror/view'

import { isValidLineRange, MOUSE_MAIN_BUTTON } from './utils'

/**
 * Represents the currently selected line range. null means no lines are
 * selected. Line numbers are 1-based.
 * endLine may be smaller than line
 */
export type SelectedLineRange = { line: number; character?: number; endLine?: number } | null

const selectedLineDecoration = Decoration.line({
    attributes: {
        tabIndex: '-1',
        'data-line-focusable': '',
        'data-testid': 'selected-line',
    },
})
const selectedLineGutterMarker = new (class extends GutterMarker {
    public elementClass = 'selected-line'
})()
export const setSelectedLines = StateEffect.define<SelectedLineRange>()
const setEndLine = StateEffect.define<number>()

/**
 * This field stores the selected line range and provides the corresponding line
 * decorations.
 */
export const selectedLines = StateField.define<SelectedLineRange>({
    create() {
        return null
    },
    compare(a, b) {
        if (a === b) {
            return true
        }
        if (!a || !b) {
            return false
        }
        return a.line === b.line && a.endLine === b.endLine
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

        /**
         * We highlight selected lines using layer instead of line decorations.
         * With this approach both selected lines and editor selection layers are be visible (with the latter taking precedence).
         *
         * With line decorations the editor selection layer is positioned behind the document content
         * and thus the line background set by line decorations overrides the layer background making selected text
         * not highlighted. An alternative would be to use reduce the opacity of the line decorations but this would
         * change the text selection color.
         */
        layer({
            above: false,
            markers(view) {
                const range = view.state.field(field)
                if (!range) {
                    return []
                }

                // We can't use RectangleMarker.fromRange because this positions the marker exactly at the start/top of
                // the actual text in the line, not at start/top what we consider to be the line. This is especially
                // apparent when the blame column is visible because the line hight is increased to give the blame
                // information more space. The following box illustrates the problem:
                //
                // ┌───┬────────────────────────────────────────────────┐ ──┐
                // │   │               ──┐                              │   │
                // │ 1 │ Some text here  ├── .fromRange gives us this   │   ├─── we want this
                // │   │               ──┘                              │   │
                // └───┴────────────────────────────────────────────────┘ ──┘
                //
                // The left and top positions, and width and height of the rectangle marker that corresponds to the
                // selected line are computed from these values:
                //
                //                rectangle.left
                //                     ◄─►    rectangle.width
                //                        ◄──────────────────────►
                //                   ▲ ┌─────────────────────────┐     ▲
                //                   │ │                         │     │documentPadding.top
                //                   │ ├─┬───────────────────────┤ ▲ ▲ ▼
                //      rectangle.top│ │1│First line             │ │ │
                //                   │ ├─┼───────────────────────┤ │ │topLineBlock.top
                //                   │ │2│Second line            │ │ │
                //                  ▲▼ ├─┼───────────────────────┤ │ ▼
                //  rectangle.height│  │3│███████████████████████│ │topLineBlock.bottom
                //                  ▼  ├─┼───────────────────────┤ ▼
                //                     │4│Fourth line            │
                //                     ├─┼───────────────────────┤
                //                     │.│...                    │
                //                     └─┴───────────────────────┘
                //                       ◄───────────────────────►
                //                     ▲ ▲   contentRect.width
                //                     │ │
                //                     │ │
                //                     │ contentRect.left
                //                     │
                //                     viewRect.left

                const viewRect = view.dom.getBoundingClientRect()
                const contentRect = view.contentDOM.getBoundingClientRect()

                const topLine = view.state.doc.line(range.endLine ? Math.min(range.line, range.endLine) : range.line)
                const topLineBlock = view.lineBlockAt(topLine.from)

                // Markers are positioned relative to the view DOM element, i.e. position 0 would be left of the gutter.
                // This computes the left position of the content element relative to the view DOM element.
                const left = contentRect.left - viewRect.left

                // block.top is relative to the document top, which is the top of the first line, _not_ including the
                // content element's padding. So we have to add the padding to properly align the marker with the top
                // of the line.
                const top = topLineBlock.top + view.documentPadding.top
                const width = contentRect.width
                let height = topLineBlock.bottom - topLineBlock.top

                if (range.endLine !== undefined) {
                    const bottomLine = view.state.doc.line(
                        Math.min(view.state.doc.lines, Math.max(range.line, range.endLine))
                    )
                    const bottomLineBlock = view.lineBlockAt(bottomLine.from)
                    height = bottomLineBlock.bottom - topLineBlock.top
                }

                return [new RectangleMarker('selected-line', left, top, width, height)]
            },
            update(update) {
                return (
                    update.docChanged ||
                    update.selectionSet ||
                    update.viewportChanged ||
                    update.transactions.some(transaction =>
                        transaction.effects.some(effect => effect.is(setSelectedLines) || effect.is(setEndLine))
                    )
                )
            },
            class: 'selected-lines-layer',
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
 * An annotation to indicate where a line selection is coming from.
 * Transactions that set selected lines without this annotation are assumed to be
 * "external" (e.g. from syncing with the URL).
 */
const lineSelectionSource = Annotation.define<'gutter'>()

/**
 * An annotation to indicate that we have to scroll the current selected line
 * into the view regardless of last selected line cache.
 */
export const lineScrollEnforcing = Annotation.define<'scroll-enforcing'>()

/**
 * View plugin responsible for scrolling the selected line(s) into view if/when
 * necessary.
 */
class ScrollIntoView implements PluginValue {
    private lastSelectedLines: SelectedLineRange | null = null
    constructor(private readonly view: EditorView, config: SelectableLineNumbersConfig) {
        this.lastSelectedLines = this.view.state.field(selectedLines)
        if (!config.skipInitialScrollIntoView) {
            this.scrollIntoView(this.lastSelectedLines)
        }
    }

    public update(update: ViewUpdate): void {
        const currentSelectedLines = update.state.field(selectedLines)
        const isForcedScroll = update.transactions.some(
            transaction => transaction.annotation(lineScrollEnforcing) === 'scroll-enforcing'
        )

        const hasSelectedLineChanged = isForcedScroll ? true : this.lastSelectedLines !== currentSelectedLines
        const isExternalTrigger = update.transactions.some(
            transaction => transaction.annotation(lineSelectionSource) !== 'gutter'
        )

        if (hasSelectedLineChanged && isExternalTrigger) {
            // Only scroll selected lines into view when the user isn't
            // currently selecting lines themselves (as indicated by the
            // presence of the "gutter" annotation). Otherwise, the scroll
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

const selectedLineNumberTheme = EditorView.theme({
    '.cm-lineNumbers': {
        cursor: 'pointer',
        color: 'var(--line-number-color)',

        '& .cm-gutterElement:hover': {
            textDecoration: 'underline',
        },
    },
})

interface SelectableLineNumbersConfig {
    onSelection: (range: SelectedLineRange) => void
    initialSelection: SelectedLineRange | null
    /**
     * If provided, this function will be called if the user
     * clicks anywhere in a line, not just on the line number.
     * In this case `onSelection` will be ignored.
     */
    onLineClick?: (line: number) => void

    /**
     * If set to true, the initial selection will not be scrolled into view.
     */
    skipInitialScrollIntoView?: boolean

    // todo(fkling): Refactor this logic, maybe move into separate extensions
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
        ViewPlugin.define(view => new ScrollIntoView(view, config)),
        selectedLines.init(() => config.initialSelection),
        lineNumbers({
            domEventHandlers: {
                mouseup(view, block, event) {
                    if (!config.onLineClick) {
                        return false
                    }

                    const mouseEvent = event as MouseEvent
                    if (mouseEvent.button !== MOUSE_MAIN_BUTTON) {
                        return false
                    }

                    const line = view.state.doc.lineAt(block.from).number
                    config.onLineClick(line)
                    return true
                },

                mousedown(view, block, event) {
                    if (config.onLineClick) {
                        return false
                    }

                    const mouseEvent = event as MouseEvent
                    if (mouseEvent.button !== MOUSE_MAIN_BUTTON) {
                        return false
                    }

                    const line = view.state.doc.lineAt(block.from).number
                    view.dispatch({
                        effects: mouseEvent.shiftKey ? setEndLine.of(line) : setSelectedLines.of({ line }),
                        annotations: lineSelectionSource.of('gutter'),
                        // Collapse/reset text selection
                        selection: { anchor: view.state.selection.main.anchor },
                    })

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
        selectedLineNumberTheme,
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
    // Only consider start and end line when determining whether to scroll a line into view.
    // Whether or not the character offset is valid doesn't matter in this case.
    const normalizedRange: SelectedLineRange = range ? { line: range.line, endLine: range.endLine } : range

    if (!normalizedRange || !isValidLineRange(normalizedRange, view.state.doc)) {
        return false
    }

    const from = view.lineBlockAt(view.state.doc.line(normalizedRange.line).from)
    const to = normalizedRange.endLine ? view.lineBlockAt(view.state.doc.line(normalizedRange.endLine).to) : from

    return (
        from.top + from.height >= view.scrollDOM.scrollTop + view.scrollDOM.clientHeight ||
        to.top <= view.scrollDOM.scrollTop
    )
}
