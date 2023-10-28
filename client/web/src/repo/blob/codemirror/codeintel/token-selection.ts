import { Annotation, EditorState, Extension, Prec, StateField, Transaction, TransactionSpec } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'

import { getCodeIntelAPI } from './api'

const tokenSelection = Annotation.define<boolean>()

console.log('create token selection module')

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

export const selectedToken = StateField.define<{ from: number; to: number } | null>({
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
                return range
            }
        }
        return value
    },
    provide: self => [
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
    ],
})

export function getSelectedToken(state: EditorState): { from: number; to: number } | null {
    return state.field(selectedToken)
}

export function setSelection(anchor: number): TransactionSpec {
    return { selection: { anchor }, annotations: tokenSelection.of(true), scrollIntoView: true }
}

export const selectedTokenExtension: Extension = selectedToken
