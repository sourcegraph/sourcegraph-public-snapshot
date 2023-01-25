import {
    selectCharLeft,
    selectCharRight,
    selectGroupLeft,
    selectGroupRight,
    selectLineDown,
    selectLineUp,
} from '@codemirror/commands'
import { EditorView, KeyBinding, keymap, layer, RectangleMarker } from '@codemirror/view'

import { blobPropsFacet } from '..'
import { syntaxHighlight } from '../highlight'
import { positionAtCmPosition, closestOccurrenceByCharacter } from '../occurrence-utils'
import { CodeIntelTooltip } from '../tooltips/CodeIntelTooltip'
import { LoadingTooltip } from '../tooltips/LoadingTooltip'

import { goToDefinitionAtOccurrence } from './definition'
import { getHoverTooltip } from './hover'
import { isModifierKeyHeld } from './modifier-click'
import { positionToOffset } from '../utils'
import { getCodeIntelTooltipState, setFocusedOccurrenceTooltip, selectOccurrence } from './code-intel-tooltips'

const keybindings: KeyBinding[] = [
    {
        key: 'Space',
        run(view) {
            const selected = getCodeIntelTooltipState(view, 'focus')
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
                    const current = getCodeIntelTooltipState(view, 'focus')
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
            const current = getCodeIntelTooltipState(view, 'focus')
            if (current?.tooltip instanceof CodeIntelTooltip) {
                view.dispatch({ effects: setFocusedOccurrenceTooltip.of(null) })
            }

            return true
        },
    },

    {
        key: 'Enter',
        run(view) {
            const selected = getCodeIntelTooltipState(view, 'focus')
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

function keyDownHandler(event: KeyboardEvent, view: EditorView) {
    switch (event.key) {
        case 'ArrowLeft': {
            const selectedOccurrence = getCodeIntelTooltipState(view, 'focus')?.occurrence
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
            const selectedOccurrence = getCodeIntelTooltipState(view, 'focus')?.occurrence
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
            const selectedOccurrence = getCodeIntelTooltipState(view, 'focus')?.occurrence
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
            const selectedOccurrence = getCodeIntelTooltipState(view, 'focus')?.occurrence
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
}

const selectionLayer = layer({
    above: false,
    markers(view) {
        return view.state.selection.ranges
            .map(r => (r.empty ? [] : RectangleMarker.forRange(view, 'cm-selectionBackground', r)))
            .reduce((a, b) => a.concat(b))
    },
    update(update, dom) {
        return update.docChanged || update.selectionSet || update.viewportChanged
    },
    class: 'cm-selectionLayer',
})

function selectionLayerExtension() {
    return [
        selectionLayer,
        EditorView.theme({
            // '.cm-line': {
            //     '& ::selection': { backgroundColor: 'transparent !important' },
            //     '&::selection': { backgroundColor: 'transparent !important' },
            // },
            '.cm-selectionLayer .cm-selectionBackground': {
                background: 'var(--code-selection-bg)',
            },
        }),
    ]
}

export function keyboardShortcutsExtension() {
    return [selectionLayerExtension(), keymap.of(keybindings), EditorView.domEventHandlers({ keydown: keyDownHandler })]
}
