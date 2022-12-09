import { defaultKeymap } from '@codemirror/commands'
import { KeyBinding } from '@codemirror/view'

import { blobPropsFacet } from '..'
import { syntaxHighlight } from '../highlight'
import { occurrenceAtPosition, positionAtCmPosition, closestOccurrenceByCharacter } from '../occurrence-utils'
import { CodeIntelTooltip } from '../tooltips/CodeIntelTooltip'
import { LoadingTooltip } from '../tooltips/LoadingTooltip'

import { goToDefinitionAtOccurrence } from './definition'
import { closeHover, getHoverTooltip, hoverField, showHover } from './hover'
import { isModifierKeyHeld } from './modifier-click'
import { selectOccurrence } from './selections'

const keybindings: readonly KeyBinding[] = [
    {
        key: 'Space',
        run(view) {
            const hover = view.state.field(hoverField)
            // Space toggles the codeintel tooltip so we guard against temporary
            // tooltips like "loading" or "no definition found".
            if (hover !== null && hover instanceof CodeIntelTooltip) {
                closeHover(view)
                return true
            }

            const position = view.state.selection.main.from + 1
            const spinner = new LoadingTooltip(view, position)
            getHoverTooltip(view, position)
                .then(
                    value => (value ? showHover(view, value) : closeHover(view)),
                    () => {}
                )
                .finally(() => spinner.stop())
            return true
        },
    },
    {
        key: 'Enter',
        run(view) {
            const position = positionAtCmPosition(view, view.state.selection.main.from)
            const occurrence = occurrenceAtPosition(view.state, position)
            if (!occurrence) {
                return false
            }
            const cmLine = view.state.doc.line(occurrence.range.start.line + 1)
            const cmPos = cmLine.from + occurrence.range.end.character
            const spinner = new LoadingTooltip(view, cmPos)
            goToDefinitionAtOccurrence(view, occurrence)
                .then(
                    ({ handler, url }) => {
                        if (view.state.field(isModifierKeyHeld) && url) {
                            window.open(url, '_blank')
                        } else {
                            handler(position)
                        }
                    },
                    () => {}
                )
                .finally(() => spinner.stop())
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
    {
        key: 'ArrowLeft',
        run(view) {
            const position = positionAtCmPosition(view, view.state.selection.main.from)
            const table = view.state.facet(syntaxHighlight)
            const line = position.line
            const occurrence = closestOccurrenceByCharacter(line, table, position, occurrence =>
                occurrence.range.start.isSmaller(position)
            )
            if (occurrence) {
                selectOccurrence(view, occurrence)
            }
            return true
        },
    },
    {
        key: 'ArrowRight',
        run(view) {
            const position = positionAtCmPosition(view, view.state.selection.main.from)
            const table = view.state.facet(syntaxHighlight)
            const line = position.line
            const occurrence = closestOccurrenceByCharacter(line, table, position, occurrence =>
                occurrence.range.start.isGreater(position)
            )
            if (occurrence) {
                selectOccurrence(view, occurrence)
            }
            return true
        },
    },
    {
        key: 'ArrowDown',
        run(view) {
            const position = positionAtCmPosition(view, view.state.selection.main.from)
            const table = view.state.facet(syntaxHighlight)
            for (let line = position.line + 1; line < table.lineIndex.length; line++) {
                const occurrence = closestOccurrenceByCharacter(line, table, position)
                if (occurrence) {
                    selectOccurrence(view, occurrence)
                    return true
                }
            }
            return true
        },
    },
    {
        key: 'ArrowUp',
        run(view) {
            const position = positionAtCmPosition(view, view.state.selection.main.from)
            const table = view.state.facet(syntaxHighlight)
            for (let line = position.line - 1; line >= 0; line--) {
                const occurrence = closestOccurrenceByCharacter(line, table, position)
                if (occurrence) {
                    selectOccurrence(view, occurrence)
                    return true
                }
            }
            return true
        },
    },
    {
        key: 'Escape',
        run(view) {
            closeHover(view)
            view.contentDOM.blur()
            return true
        },
    },
]

const remappedKeyBindings: Record<string, string> = {
    // Open definition in a new tab.
    Enter: 'Mod-Enter',
    // Vim bindings.
    ArrowUp: 'k',
    ArrowDown: 'j',
    ArrowLeft: 'h',
    ArrowRight: 'l',
}

export const tokenSelectionKeyBindings = [
    ...keybindings.flatMap(keybinding => {
        const remappedKey = remappedKeyBindings[keybinding.key ?? '']
        if (remappedKey) {
            return [keybinding, { ...keybinding, key: remappedKey }]
        }
        return [keybinding]
    }),
    // Fallback to the default arrow keys to allow, for example, Shift-ArrowLeft
    ...defaultKeymap.filter(({ key }) => key?.includes('Arrow')),
]
