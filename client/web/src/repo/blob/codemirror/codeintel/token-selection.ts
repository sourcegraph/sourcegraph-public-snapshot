import {
    Annotation,
    EditorState,
    Extension,
    Prec,
    StateEffect,
    StateField,
    Transaction,
    TransactionSpec,
} from '@codemirror/state'
import { Decoration, EditorView, keymap } from '@codemirror/view'
import { concat, from, of } from 'rxjs'
import { timeoutWith } from 'rxjs/operators'

import { syntaxHighlight } from '../highlight'
import { positionAtCmPosition, closestOccurrenceByCharacter } from '../occurrence-utils'
import { LoadingTooltip } from '../tooltips/LoadingTooltip'
import { positionToOffset, preciseOffsetAtCoords } from '../utils'

import { getCodeIntelAPI } from './api'
import { TooltipSource, showTooltip } from './tooltips'

const tokenSelection = Annotation.define<boolean>()

/**
 * Returns `true` if the editor selection is empty or is inside the occurrence range.
 */
function shouldApplyFocusStyles(state: EditorState, range: { from: number; to: number }): boolean {
    if (state.selection.main.empty) {
        return true
    }

    const isEditorSelectionInsideOccurrenceRange =
        state.selection.main.from >= range.from && state.selection.main.to <= range.to
    return isEditorSelectionInsideOccurrenceRange
}

const interactiveTokenDecoration = Decoration.mark({
    class: 'interactive-occurrence', // used as interactive occurrence selector
    attributes: {
        // Selected (focused) occurrence is the only focusable element in the editor.
        // This helps to maintain the focus position when editor is blurred and then focused again.
        tabindex: '0',
    },
})
const focusedTokenDecoration = Decoration.mark({ class: 'focus-visible' })

function shouldUpdateSelectedToken(transaction: Transaction): boolean {
    return transaction.isUserEvent('select.pointer') || !!transaction.annotation(tokenSelection)
}

const setTooltipSource = StateEffect.define<TooltipSource | null>()

function hideTooltip(view: EditorView): void {
    view.dispatch({ effects: setTooltipSource.of(null) })
}

export const selectedToken = StateField.define<{
    from: number
    to: number
    tooltipSource?: TooltipSource | null
} | null>({
    create() {
        return null
    },
    update(value, transaction) {
        if (shouldUpdateSelectedToken(transaction)) {
            const range = getCodeIntelAPI(transaction.state).findOccurrenceRangeAt(
                transaction.newSelection.main.from,
                transaction.state
            )
            if (range) {
                value = range
            }
        }
        if (value) {
            for (const effect of transaction.effects) {
                if (effect.is(setTooltipSource) && effect.value !== value.tooltipSource) {
                    value = { ...value, tooltipSource: effect.value }
                }
            }
        }
        return value
    },
    provide: self => [
        /**
         * Register tooltip source when available.
         */
        showTooltip.computeN([self], state => {
            const field = state.field(self)
            return field?.tooltipSource ? [{ range: field, key: 'selection', source: field.tooltipSource }] : []
        }),

        /**
         * This extension is responsible for keeping the focus inside the editor so
         * that keyboard navigation continues to work.
         *
         * Because CodeMirror only renders the visible part of the document it's
         * possible that the line with the selected token gets removed. If the token
         * was focused the whole editor would loose focus.
         *
         * To prevent this, this extension ensures focus in the following situations:
         *   - a new token is selected -> decoration/token itself is focused
         *   - token is scrolled out of view -> editor content is selected
         */
        EditorView.updateListener.of(update => {
            const range = update.state.field(self)
            const view = update.view
            if (range) {
                if (range !== update.startState.field(self)) {
                    view.contentDOM.querySelector<HTMLElement>('.interactive-occurrence')?.focus()
                }
                if (
                    update.viewportChanged &&
                    !view.dom.contains(document.activeElement) &&
                    (!view.contentDOM.querySelector<HTMLElement>('.interactive-occurrence') ||
                        range.from < view.viewport.from)
                ) {
                    view.contentDOM.focus()
                }
            }
        }),

        // Prec.lowest causes the decoration to wrap around any other decoration that's inside the token,
        // ensuring that it's not broken up across multiple decorations (e.g. by search highlights)
        Prec.lowest(
            EditorView.decorations.compute([self, 'selection'], state => {
                const selectedRange = state.field(self)
                if (!selectedRange) {
                    return Decoration.none
                }
                const decorations: Decoration[] = [interactiveTokenDecoration]
                if (shouldApplyFocusStyles(state, selectedRange)) {
                    decorations.unshift(focusedTokenDecoration)
                }
                return Decoration.set(decorations.map(d => d.range(selectedRange.from, selectedRange.to)))
            })
        ),

        /**
         * If there is a focused occurrence set editor's tabindex to -1, so that pressing Shift+Tab moves the focus
         * outside the editor instead of focusing the editor itself.
         *
         * Explicitly define extension precedence to override the [default tabindex value](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@728ea45d1cc063cd60cbd552e00929c09cb8ced8/-/blob/client/web/src/repo/blob/CodeMirrorBlob.tsx?L47&).
         */
        Prec.high(
            EditorView.contentAttributes.compute([self], state => ({
                tabindex: state.field(self) ? '-1' : '0',
            }))
        ),

        /**
         * Keybindings for showing/hiding a code intel toolip at this position
         */
        keymap.of([
            {
                key: 'Space',
                run(view) {
                    const selected = view.state.field(self)
                    if (!selected) {
                        return true
                    }

                    if (selected.tooltipSource) {
                        hideTooltip(view)
                        return true
                    }

                    const tooltip$ = from(getCodeIntelAPI(view.state).getHoverTooltip(view.state, selected))
                    view.dispatch({
                        effects: setTooltipSource.of(
                            tooltip$.pipe(
                                timeoutWith(50, concat(of(new LoadingTooltip(selected.from, selected.to)), tooltip$))
                            )
                        ),
                    })
                    return true
                },
            },
            {
                key: 'Escape',
                run(view) {
                    view.dispatch({ effects: setTooltipSource.of(null) })
                    return true
                },
            },

            {
                key: 'Enter',
                run(view) {
                    const selected = getSelectedToken(view.state)
                    if (!selected) {
                        return false
                    }

                    view.dispatch({ effects: setTooltipSource.of(new LoadingTooltip(selected.from, selected.to)) })
                    getCodeIntelAPI(view.state)
                        .goToDefinitionAt(view, selected.from)
                        .finally(() => {
                            hideTooltip(view)
                        })
                    return true
                },
            },
            {
                key: 'Mod-Enter',
                run(view) {
                    const selected = getSelectedToken(view.state)
                    if (!selected) {
                        return false
                    }

                    view.dispatch({ effects: setTooltipSource.of(new LoadingTooltip(selected.from, selected.to)) })
                    getCodeIntelAPI(view.state)
                        .goToDefinitionAt(view, selected.from, { newWindow: true })
                        .finally(() => {
                            hideTooltip(view)
                        })
                    return true
                },
            },
        ]),

        EditorView.domEventHandlers({
            /**
             * Keyboard event handlers defined via {@link keymap} facet do not work with the screen reader enabled while
             * keypress handlers defined via {@link EditorView.domEventHandlers} still work.
             */
            keydown: (event: KeyboardEvent, view: EditorView): boolean => {
                switch (event.key) {
                    case 'ArrowLeft': {
                        const from = view.state.field(self)?.from || view.viewport.from
                        const position = positionAtCmPosition(view.state.doc, from)
                        const table = view.state.facet(syntaxHighlight)
                        const occurrence = closestOccurrenceByCharacter(position.line, table, position, occurrence =>
                            occurrence.range.start.isSmaller(position)
                        )
                        const anchor = occurrence ? positionToOffset(view.state.doc, occurrence.range.start) : null
                        if (anchor !== null) {
                            view.dispatch(setSelection(anchor))
                        }

                        return true
                    }
                    case 'ArrowRight': {
                        const from = view.state.field(self)?.from || view.viewport.from
                        const position = positionAtCmPosition(view.state.doc, from)
                        const table = view.state.facet(syntaxHighlight)
                        const occurrence = closestOccurrenceByCharacter(position.line, table, position, occurrence =>
                            occurrence.range.start.isGreater(position)
                        )
                        const anchor = occurrence ? positionToOffset(view.state.doc, occurrence.range.start) : null
                        if (anchor !== null) {
                            view.dispatch(setSelection(anchor))
                        }

                        return true
                    }
                    case 'ArrowUp': {
                        const from = view.state.field(self)?.from || view.viewport.from
                        const position = positionAtCmPosition(view.state.doc, from)
                        const table = view.state.facet(syntaxHighlight)
                        for (let line = position.line - 1; line >= 0; line--) {
                            const occurrence = closestOccurrenceByCharacter(line, table, position)
                            const anchor = occurrence ? positionToOffset(view.state.doc, occurrence.range.start) : null
                            if (anchor !== null) {
                                view.dispatch(setSelection(anchor))
                                return true
                            }
                        }
                        return true
                    }
                    case 'ArrowDown': {
                        const from = view.state.field(self)?.from || view.viewport.from
                        const position = positionAtCmPosition(view.state.doc, from)
                        const table = view.state.facet(syntaxHighlight)
                        for (let line = position.line + 1; line < table.lineIndex.length; line++) {
                            const occurrence = closestOccurrenceByCharacter(line, table, position)
                            const anchor = occurrence ? positionToOffset(view.state.doc, occurrence.range.start) : null
                            if (anchor !== null) {
                                view.dispatch(setSelection(anchor))
                                return true
                            }
                        }
                        return true
                    }

                    default: {
                        return false
                    }
                }
            },

            /**
             * This extension closes focus-related tooltips when selection moves elsewhere.
             */
            click(event, view) {
                const offset = preciseOffsetAtCoords(view, event)
                if (offset) {
                    const range = getCodeIntelAPI(view.state).findOccurrenceRangeAt(offset, view.state)
                    if (range && range?.from === view.state.field(self)?.from) {
                        return
                    }
                }
                hideTooltip(view)
            },
        }),
    ],
})

export function getSelectedToken(state: EditorState): { from: number; to: number } | null {
    return state.field(selectedToken)
}

export function setSelection(anchor: number): TransactionSpec {
    return { selection: { anchor }, annotations: tokenSelection.of(true), scrollIntoView: true }
}

export const selectedTokenExtension: Extension = selectedToken
