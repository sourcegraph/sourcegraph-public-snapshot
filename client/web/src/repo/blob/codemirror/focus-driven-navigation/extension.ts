import {
    Decoration,
    EditorView,
    getTooltip,
    KeyBinding,
    keymap,
    PluginValue,
    showTooltip,
    Tooltip,
    ViewPlugin,
} from '@codemirror/view'
import { Extension, StateEffect, StateField } from '@codemirror/state'
import { isSelectionInsideDocument, positionToOffset, preciseOffsetAtCoords, sortRangeValuesByStart } from '../utils'
import {
    closestOccurrenceByCharacter,
    occurrenceAtPosition,
    positionAtCmPosition,
    rangeToCmSelection,
} from '../occurrence-utils'
import { Occurrence } from '@sourcegraph/shared/src/codeintel/scip'
import { fallbackOccurrences, selectionFromLocation } from '../token-selection/selections'
import { syntaxHighlight } from '../highlight'
import {
    documentHighlightsExtension,
    showDocumentHighlightsForOccurrence,
} from '../token-selection/document-highlights'
import { showDocumentHighlights } from '../document-highlights'
import { closeHover, getHoverTooltip, hoverCache } from '../token-selection/hover'
import { definitionCache, goToDefinitionAtOccurrence, underlinedDefinitionFacet } from '../token-selection/definition'
import { from, fromEvent, of, Subscription } from 'rxjs'
import { catchError, debounceTime, filter, map, scan, switchMap, tap } from 'rxjs/operators'
import { computeMouseDirection, HOVER_DEBOUNCE_TIME, MOUSE_NO_BUTTON } from '../hovercard'
import { CodeIntelTooltip } from '../tooltips/CodeIntelTooltip'
import {
    selectCharLeft,
    selectCharRight,
    selectGroupLeft,
    selectGroupRight,
    selectLineDown,
    selectLineUp,
} from '@codemirror/commands'
import { blobPropsFacet } from '../index'
import { isModifierKeyHeld, modifierClickFacet } from '../token-selection/modifier-click'
import * as H from 'history'

const setFocusedOccurrence = StateEffect.define<Occurrence | null>()
const setHoveredOccurrence = StateEffect.define<{ occurrence: Occurrence; tooltip: Tooltip | null } | null>()

const setFocusedOccurrenceTooltip = StateEffect.define<Tooltip | null>()

export const selectedOccurrenceField = StateField.define<{
    hover: { occurrence: Occurrence; tooltip: Tooltip | null } | null
    focus: { occurrence: Occurrence; tooltip: Tooltip | null } | null
}>({
    create() {
        return { hover: null, focus: null }
    },
    update(value, transaction) {
        if (transaction.selection) {
            console.log(transaction.selection.main)
        }
        for (const effect of transaction.effects) {
            if (effect.is(setFocusedOccurrence)) {
                return {
                    ...value,
                    focus: effect.value ? { occurrence: effect.value, tooltip: null } : null,
                }
            }
            if (effect.is(setFocusedOccurrenceTooltip)) {
                // TODO: check if tooltip pos matches selected occurrence
                return {
                    ...value,
                    focus: value.focus?.occurrence ? { ...value.focus, tooltip: effect.value } : null,
                }
            }

            if (effect.is(setHoveredOccurrence)) {
                return { ...value, hover: effect.value }
            }
        }
        return value
    },
    provide(field) {
        return [
            EditorView.decorations.compute([field, showDocumentHighlights], state => {
                const { focus, hover } = state.field(field)

                // TODO: add support for hovered occurrence highlights
                const highlights = state.facet(showDocumentHighlights)

                const ranges = []

                if (focus) {
                    const valueRangeStart = positionToOffset(state.doc, focus.occurrence.range.start)
                    const valueRangeEnd = positionToOffset(state.doc, focus.occurrence.range.end)

                    if (valueRangeStart !== null && valueRangeEnd !== null) {
                        ranges.push(
                            Decoration.mark({
                                class: 'interactive-occurrence sourcegraph-document-highlight',
                                attributes: { tabindex: '0' },
                            }).range(valueRangeStart, valueRangeEnd)
                        )
                    }

                    const selected = highlights?.find(
                        ({ range }) =>
                            focus.occurrence.range.start.line === range.start.line &&
                            focus.occurrence.range.start.character === range.start.character &&
                            focus.occurrence.range.end.line === range.end.line &&
                            focus.occurrence.range.end.character === range.end.character
                    )

                    if (selected) {
                        for (const highlight of sortRangeValuesByStart(highlights)) {
                            if (highlight === selected) {
                                continue
                            }

                            const highlightRangeStart = positionToOffset(state.doc, highlight.range.start)
                            const highlightRangeEnd = positionToOffset(state.doc, highlight.range.end)

                            if (highlightRangeStart === null || highlightRangeEnd === null) {
                                continue
                            }

                            ranges.push(
                                Decoration.mark({
                                    class: 'interactive-occurrence sourcegraph-document-highlight',
                                }).range(highlightRangeStart, highlightRangeEnd)
                            )
                        }
                    }
                }

                if (hover && hover.occurrence !== focus?.occurrence) {
                    const valueRangeStart = positionToOffset(state.doc, hover.occurrence.range.start)
                    const valueRangeEnd = positionToOffset(state.doc, hover.occurrence.range.end)

                    if (valueRangeStart !== null && valueRangeEnd !== null) {
                        ranges.push(
                            Decoration.mark({
                                class: 'interactive-occurrence selection-highlight',
                            }).range(valueRangeStart, valueRangeEnd)
                        )
                    }
                }

                return Decoration.set(ranges.sort((a, b) => a.from - b.from))
            }),

            showTooltip.computeN([field], state => {
                const { hover, focus } = state.field(field)
                const tooltips = []
                if (focus?.tooltip) {
                    tooltips.push(focus.tooltip)
                }
                if (hover?.tooltip && hover.occurrence !== focus?.occurrence) {
                    tooltips.push(hover.tooltip)
                }
                return tooltips
            }),
        ]
    },
})

const focusOccurrence = (view: EditorView, occurrence: Occurrence): void => {
    const offset = positionToOffset(view.state.doc, occurrence.range.end)
    if (offset !== null) {
        const node = view.domAtPos(offset).node
        const el = node instanceof HTMLElement ? node : node.parentElement
        const lineEl = el?.matches('.cm-line') ? el : el?.closest('.cm-line')
        const interactiveOccurrenceEl = lineEl?.querySelector<HTMLElement>('.interactive-occurrence')
        if (interactiveOccurrenceEl) {
            interactiveOccurrenceEl?.focus()
        }
    }
}

export const selectOccurrence = (view: EditorView, occurrence: Occurrence): void => {
    // TODO: warmup occurrence
    view.dispatch({
        effects: setFocusedOccurrence.of(occurrence),
        selection: rangeToCmSelection(view.state, occurrence.range),
    })
    showDocumentHighlightsForOccurrence(view, occurrence)
    focusOccurrence(view, occurrence)
}

class LoadingTooltip implements Tooltip {
    public readonly above = true

    constructor(public readonly pos: number) {}

    public create() {
        const dom = document.createElement('div')
        dom.classList.add('tmp-tooltip')
        dom.textContent = 'Loading...'
        return { dom }
    }
}

const keybindings: KeyBinding[] = [
    {
        key: 'Space',
        run(view) {
            const selected = view.state.field(selectedOccurrenceField).focus
            if (!selected) {
                return true
            }

            if (selected.tooltip instanceof CodeIntelTooltip) {
                view.dispatch({ effects: setFocusedOccurrenceTooltip.of(null) })
                return true
            }

            const offset = positionToOffset(view.state.doc, selected.occurrence.range.start)
            if (offset === null) {
                return true
            }

            // show loading tooltip
            view.dispatch({ effects: setFocusedOccurrenceTooltip.of(new LoadingTooltip(offset)) })

            getHoverTooltip(view, offset)
                .then(value => view.dispatch({ effects: setFocusedOccurrenceTooltip.of(value) }))
                .finally(() => {
                    // close loading tooltip if any
                    const current = view.state.field(selectedOccurrenceField).focus
                    if (current?.tooltip instanceof LoadingTooltip && current?.occurrence === selected.occurrence) {
                        view.dispatch({ effects: setFocusedOccurrenceTooltip.of(null) })
                    }
                })

            return true
        },
    },
    {
        key: 'Escape',
        run(view) {
            const current = view.state.field(selectedOccurrenceField).focus
            if (current?.tooltip instanceof CodeIntelTooltip) {
                view.dispatch({ effects: setFocusedOccurrenceTooltip.of(null) })
            }

            return true
        },
    },

    {
        key: 'Enter',
        run(view) {
            const selected = view.state.field(selectedOccurrenceField).focus
            if (!selected?.occurrence) {
                return false
            }

            const offset = positionToOffset(view.state.doc, selected.occurrence.range.start)
            if (offset === null) {
                return true
            }

            // show loading tooltip
            view.dispatch({ effects: setFocusedOccurrenceTooltip.of(new LoadingTooltip(offset)) })

            goToDefinitionAtOccurrence(view, selected.occurrence)
                .then(
                    ({ handler, url }) => {
                        if (view.state.field(isModifierKeyHeld) && url) {
                            window.open(url, '_blank')
                        } else {
                            handler(selected.occurrence.range.start)
                        }
                    },
                    () => {}
                )
                .finally(() => {
                    // hide loading tooltip
                    view.dispatch({ effects: setFocusedOccurrenceTooltip.of(null) })
                })

            // TODO: go to definition at occurrence
            return true
        },
    },

    {
        key: 'Mod-ArrowRight',
        run(view) {
            view.state.facet(blobPropsFacet).history.goForward()
            return true
        },
    },
    {
        key: 'Mod-ArrowLeft',
        run(view) {
            view.state.facet(blobPropsFacet).history.goBack()
            return true
        },
    },

    // TODO: window selection is not updated when manually setting CodeMirror selection.
    // Check why updateSelection is not called in this case: https://sourcegraph.com/github.com/codemirror/view@fd097ac61a3ca0b3b6f6ea958d04071ecaf7c231/-/blob/src/docview.ts?L146:3-146:18#tab=references
    // but is called when we select occurrence (via `view.dispatch({selection})`).

    {
        key: 'ArrowLeft',
        shift: selectCharLeft,
    },

    {
        key: 'ArrowRight',
        shift: selectCharRight,
    },

    {
        key: 'Mod-ArrowLeft',
        mac: 'Alt-ArrowLeft',
        shift: selectGroupLeft,
    },

    {
        key: 'Mod-ArrowRight',
        mac: 'Alt-ArrowRight',
        shift: selectGroupRight,
    },

    {
        key: 'ArrowUp',
        shift: selectLineUp,
    },

    {
        key: 'ArrowDown',
        shift: selectLineDown,
    },
]

const domEventHandlers = EditorView.domEventHandlers({
    keydown(event, view) {
        switch (event.key) {
            case 'ArrowLeft': {
                const selectedOccurrence = view.state.field(selectedOccurrenceField)?.focus?.occurrence
                const position = selectedOccurrence?.range.start || positionAtCmPosition(view, view.viewport.from)
                const table = view.state.facet(syntaxHighlight)
                const occurrence = closestOccurrenceByCharacter(position.line, table, position, occurrence =>
                    occurrence.range.start.isSmaller(position)
                )
                if (occurrence) {
                    selectOccurrence(view, occurrence)
                }

                return true
            }
            case 'ArrowRight': {
                const selectedOccurrence = view.state.field(selectedOccurrenceField)?.focus?.occurrence
                const position = selectedOccurrence?.range.start || positionAtCmPosition(view, view.viewport.from)
                const table = view.state.facet(syntaxHighlight)
                const occurrence = closestOccurrenceByCharacter(position.line, table, position, occurrence =>
                    occurrence.range.start.isGreater(position)
                )
                if (occurrence) {
                    selectOccurrence(view, occurrence)
                }

                return true
            }
            case 'ArrowUp': {
                const selectedOccurrence = view.state.field(selectedOccurrenceField)?.focus?.occurrence
                const position = selectedOccurrence?.range.start || positionAtCmPosition(view, view.viewport.from)
                const table = view.state.facet(syntaxHighlight)
                for (let line = position.line - 1; line >= 0; line--) {
                    const occurrence = closestOccurrenceByCharacter(line, table, position)
                    if (occurrence) {
                        selectOccurrence(view, occurrence)
                        return true
                    }
                }
                return true
            }
            case 'ArrowDown': {
                const selectedOccurrence = view.state.field(selectedOccurrenceField)?.focus?.occurrence
                const position = selectedOccurrence?.range.start || positionAtCmPosition(view, view.viewport.from)
                const table = view.state.facet(syntaxHighlight)
                for (let line = position.line + 1; line < table.lineIndex.length; line++) {
                    const occurrence = closestOccurrenceByCharacter(line, table, position)
                    if (occurrence) {
                        selectOccurrence(view, occurrence)
                        return true
                    }
                }
                return true
            }

            default:
                return false
        }
    },
})

/**
 * Listens to mousemove events, determines whether the position under the mouse
 * cursor is a valid {@link Occurrence}, fetches hover information as necessary and updates {@link hoverTooltip}.
 */
const hoverManager = ViewPlugin.fromClass(
    class HoverManager implements PluginValue {
        private subscription: Subscription = new Subscription()

        constructor(private readonly view: EditorView) {
            this.subscription.add(
                fromEvent<MouseEvent>(this.view.dom, 'mousemove')
                    .pipe(
                        // Debounce events so that users can move over tokens without triggering hovercards immediately
                        debounceTime(HOVER_DEBOUNCE_TIME),

                        // Ignore some events
                        filter(event => {
                            // Ignore events when hovering over an existing hovercard.
                            // This causes existing hovercards to stay open.
                            if (
                                (event.target as HTMLElement | null)?.closest(
                                    '.cm-code-intel-hovercard:not(.cm-code-intel-hovercard-pinned)'
                                )
                            ) {
                                return false
                            }

                            // We have to forward any move events that also have a
                            // button pressed. User is probably selecting text and
                            // hovercards should be hidden.
                            if (event.buttons !== MOUSE_NO_BUTTON) {
                                return true
                            }

                            // Ignore events inside the current hover range. Without this
                            // hovercards flicker when the active range is wider than the
                            // word-under-cursor range. For example, hovering over
                            //
                            // import ( "io/fs" )
                            //
                            // will detect `io` and `fs` as separate words (and would
                            // therefore trigger two individual word lookups), but the
                            // hover information returned by the server is for the whole
                            // `io/fs` range.
                            const offset = preciseOffsetAtCoords(view, event)
                            if (offset === null) {
                                return true
                            }

                            const currentOccurrence = view.state.field(selectedOccurrenceField)?.hover?.occurrence
                            if (!currentOccurrence) {
                                return true
                            }

                            return !isOffsetInHoverRange(
                                offset,
                                rangeToCmSelection(this.view.state, currentOccurrence.range)
                            )
                        }),

                        // To make it easier to reach the tooltip with the mouse, we determine
                        // in which direction the mouse moves and only hide the tooltip when
                        // the mouse moves away from it.
                        scan(
                            (
                                previous: {
                                    x: number
                                    y: number
                                    target: EventTarget | null
                                    buttons: number
                                    direction?: 'towards' | 'away' | undefined
                                },
                                next
                            ) => {
                                const currentTooltip = view.state.field(selectedOccurrenceField)?.hover?.tooltip
                                if (!currentTooltip) {
                                    return next
                                }

                                const tooltipView = getTooltip(view, currentTooltip)
                                if (!tooltipView) {
                                    return next
                                }

                                const direction = computeMouseDirection(
                                    tooltipView.dom.getBoundingClientRect(),
                                    previous,
                                    next
                                )
                                return { x: next.x, y: next.y, buttons: next.buttons, target: next.target, direction }
                            }
                        ),

                        // Determine the precise location of the word under the cursor.
                        switchMap(position => {
                            // Hide any tooltip when
                            // - the mouse is over an element that is not part of
                            //   the content. This seems necessary to make tooltips
                            //   not appear and hide open tooltips when the mouse
                            //   moves over the editor's search panel.
                            // - the user starts to select text
                            if (
                                position.buttons !== MOUSE_NO_BUTTON ||
                                !position.target ||
                                !this.view.contentDOM.contains(position.target as Node)
                            ) {
                                return of('HIDE' as const)
                            }

                            const offset = preciseOffsetAtCoords(this.view, position)
                            if (!offset) {
                                return of(null)
                            }
                            const pos = positionAtCmPosition(view, offset)
                            return of(occurrenceAtPosition(this.view.state, pos)).pipe(
                                tap(occurrence => {
                                    const current = this.view.state.field(selectedOccurrenceField)?.hover?.occurrence
                                    if (current && occurrence && current !== occurrence) {
                                        view.dispatch({
                                            effects: setHoveredOccurrence.of(null),
                                        })
                                    }
                                }),
                                switchMap(occurrence => {
                                    if (!occurrence) {
                                        return of(null)
                                    }

                                    const offset = positionToOffset(this.view.state.doc, occurrence.range.start)
                                    if (offset === null) {
                                        return of(null)
                                    }

                                    // show loading tooltip
                                    view.dispatch({
                                        effects: setHoveredOccurrence.of({
                                            occurrence,
                                            tooltip: new LoadingTooltip(offset),
                                        }),
                                    })

                                    return from(getHoverTooltip(view, offset)).pipe(
                                        catchError(() => of(null)),

                                        // close loading tooltip
                                        tap(() => {
                                            const current = view.state.field(selectedOccurrenceField).hover
                                            if (
                                                current?.tooltip instanceof LoadingTooltip &&
                                                current?.occurrence === occurrence
                                            ) {
                                                view.dispatch({ effects: setHoveredOccurrence.of(null) })
                                            }
                                        }),

                                        map(tooltip => (tooltip ? { tooltip, occurrence } : null))
                                    )
                                }),
                                map(hover => ({ position, hover }))
                            )
                        })
                    )
                    .subscribe(next => {
                        if (next === 'HIDE') {
                            view.dispatch({ effects: setHoveredOccurrence.of(null) })
                            return
                        }

                        // We only change the tooltip when
                        // a) There is a new tooltip at the position (tooltip !== null)
                        // b) there is no tooltip and the mouse is moving away from the tooltip
                        if (next?.hover || next?.position.direction !== 'towards') {
                            if (!next?.hover?.occurrence) {
                                view.dispatch({
                                    effects: setHoveredOccurrence.of(null),
                                })
                                return
                            }

                            view.dispatch({
                                effects: setHoveredOccurrence.of(next.hover),
                            })
                        }
                    })
            )
        }

        public destroy(): void {
            this.subscription.unsubscribe()
        }
    }
)

function isOffsetInHoverRange(offset: number, range: { from: number; to: number }): boolean {
    return range.from <= offset && offset <= range.to
}

// View plugin that listens to history location changes and updates editor
// selection accordingly.
const syncSelectionWithURL: Extension = ViewPlugin.fromClass(
    class implements PluginValue {
        private onDestroy: H.UnregisterCallback
        constructor(public view: EditorView) {
            const history = view.state.facet(blobPropsFacet).history
            this.onDestroy = history.listen(location => this.onLocation(location))
        }
        public onLocation(location: H.Location): void {
            const { selection } = selectionFromLocation(this.view, location)
            if (selection && isSelectionInsideDocument(selection, this.view.state.doc)) {
                const position = positionAtCmPosition(this.view, selection.from)
                const occurrence = occurrenceAtPosition(this.view.state, position)
                if (occurrence) {
                    selectOccurrence(this.view, occurrence)
                }
            }
        }
        public destroy(): void {
            this.onDestroy()
        }
    }
)

export function focusDrivenCodeNavigation() {
    return [
        documentHighlightsExtension(),
        selectedOccurrenceField,

        syncSelectionWithURL,
        modifierClickFacet.of(false),
        fallbackOccurrences,
        hoverCache,
        definitionCache,
        underlinedDefinitionFacet.of(null),

        hoverManager,
        keymap.of(keybindings),
        domEventHandlers,
    ]
}
