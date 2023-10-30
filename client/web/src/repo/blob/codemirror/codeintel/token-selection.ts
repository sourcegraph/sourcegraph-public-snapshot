import {
    Annotation,
    EditorState,
    Extension,
    Prec,
    StateEffect,
    StateField,
    Transaction,
    TransactionSpec,
    Range,
} from '@codemirror/state'
import { Decoration, EditorView, keymap } from '@codemirror/view'
import { LineOrPositionOrRange } from 'node_modules/@sourcegraph/common/out/src'
import { concat, from, of } from 'rxjs'
import { timeoutWith } from 'rxjs/operators'

import { LoadingTooltip } from '../tooltips/LoadingTooltip'
import { lprToRange } from '../utils'

import {
    findOccurrenceRangeAt,
    getHoverTooltip,
    goToDefinitionAt,
    nextOccurrencePosition,
    prevOccurrencePosition,
} from './api'
import { ignoreDecorations } from './decorations'
import { showDocumentHighlights } from './document-highlights'
import { TooltipSource, showTooltip } from './tooltips'

const tokenSelection = Annotation.define<boolean>()
const setTooltipSource = StateEffect.define<TooltipSource | null>()

const interactiveTokenDeco = Decoration.mark({
    class: 'interactive-occurrence', // used as interactive occurrence selector
    attributes: {
        // Selected (focused) occurrence is the only focusable element in the editor.
        // This helps to maintain the focus position when editor is blurred and then focused again.
        tabindex: '0',
    },
})
const focusedTokenDeco = Decoration.mark({ class: 'focus-visible' })

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

/**
 * Helper function for hiding the tooltip at the selected token.
 */
function hideTooltip(view: EditorView): void {
    view.dispatch({ effects: setTooltipSource.of(null) })
}

/**
 * Returns true if the transaction is a click or an explicit token navigation transaction.
 */
function shouldUpdateSelectedToken(transaction: Transaction): boolean {
    return transaction.isUserEvent('select.pointer') || !!transaction.annotation(tokenSelection)
}

/**
 * Field for keeping track of the currently selected token for token navigation. The
 * field updates when the selection changes to a different token via clicks or explicit
 * token navigation events.
 * It also provides
 *   - keyboard events for showing tooltips via {@link showTooltip}.
 *   - document highlights for the selected token via {@link showDocumentHighlights}
 */
const selectedToken = StateField.define<{
    range: { from: number; to: number }
    tooltipSource?: TooltipSource | null
} | null>({
    create(state) {
        const offset = state.selection.main.from
        const range = findOccurrenceRangeAt(state, offset)
        return range ? { range } : null
    },
    update(value, transaction) {
        if (shouldUpdateSelectedToken(transaction)) {
            const offset = transaction.newSelection.main.from
            if (!value || offset < value.range.from || value.range.to < offset) {
                const range = findOccurrenceRangeAt(transaction.state, offset)
                if (range) {
                    value = { range }
                } else if (value) {
                    value = { ...value, tooltipSource: null }
                }
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
            return field?.tooltipSource ? [{ range: field.range, source: field.tooltipSource }] : []
        }),

        /**
         * Show document highlights for selected token.
         */
        showDocumentHighlights.computeN([self], state => {
            const field = state.field(self)
            return field ? [field.range] : []
        }),

        /**
         * We can't add/remove any decorations inside the selected token, because
         * that causes the node to be recreated and loosing focus, which breaks
         * token keyboard navigation.
         * This facet should be used by all codeIntel extensions to ensure that any
         * conflicting decoration is removed.
         */
        ignoreDecorations.from(self, value => value?.range ?? null),

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
            const range = update.state.field(self)?.range
            const view = update.view
            if (range) {
                if (range !== update.startState.field(self)?.range) {
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
                const selectedRange = state.field(self)?.range
                if (!selectedRange) {
                    return Decoration.none
                }
                const decorations: Range<Decoration>[] = [
                    interactiveTokenDeco.range(selectedRange.from, selectedRange.to),
                ]
                if (shouldApplyFocusStyles(state, selectedRange)) {
                    decorations.unshift(focusedTokenDeco.range(selectedRange.from, selectedRange.to))
                }
                return Decoration.set(decorations, true)
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

                    const tooltip$ = from(getHoverTooltip(view.state, selected.range.from))
                    view.dispatch({
                        effects: setTooltipSource.of(
                            tooltip$.pipe(
                                timeoutWith(
                                    50,
                                    concat(of(new LoadingTooltip(selected.range.from, selected.range.to)), tooltip$)
                                )
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
                    goToDefinitionAt(view, selected.from).finally(() => {
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
                    goToDefinitionAt(view, selected.from, { newWindow: true }).finally(() => {
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
                const from = view.state.field(self)?.range.from || view.viewport.from
                let offset: number | null = null

                switch (event.key) {
                    case 'ArrowLeft': {
                        offset = prevOccurrencePosition(view.state, from, 'character')
                        break
                    }
                    case 'ArrowRight': {
                        offset = nextOccurrencePosition(view.state, from, 'character')
                        break
                    }
                    case 'ArrowUp': {
                        offset = prevOccurrencePosition(view.state, from, 'line')
                        break
                    }
                    case 'ArrowDown': {
                        offset = nextOccurrencePosition(view.state, from, 'line')
                        break
                    }

                    default: {
                        return false
                    }
                }
                if (offset !== null) {
                    view.dispatch(setSelection(offset))
                }
                return true
            },
        }),
    ],
})

function setSelection(anchor: number): TransactionSpec {
    return { selection: { anchor }, annotations: tokenSelection.of(true), scrollIntoView: true }
}

/**
 * Returns the currently focused/selected token, if any.
 */
export function getSelectedToken(state: EditorState): { from: number; to: number } | null {
    return state.field(selectedToken)?.range ?? null
}

/**
 * Helper function for moving token selection to the provided position.
 */
export function syncSelection(view: EditorView, position: LineOrPositionOrRange): void {
    const { line, character, endCharacter, endLine } = position
    if (line && character && endCharacter && (!endLine || line === endLine)) {
        const range = lprToRange(view.state.doc, { line, character, endCharacter, endLine })
        if (range) {
            view.dispatch(setSelection(range.from))
        }
    }
}

export const selectedTokenExtension: Extension = selectedToken
