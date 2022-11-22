import { KeyBinding } from '@codemirror/view'

import { syntaxHighlight } from './highlight'
import { Spinner } from './Spinner'
import { occurrenceAtPosition, positionAtCmPosition, closestOccurrence } from './positions'
import { getHoverTooltip, hoverField, setHoverEffect } from './textdocument-hover'
import { historyFacet } from './textdocument-facets'
import { selectOccurrence } from './textdocument-selections'
import { goToDefinitionAtOccurrence } from './textdocument-definition'

const keybindings: readonly KeyBinding[] = [
    {
        key: 'Space',
        run(view) {
            const hover = view.state.field(hoverField)
            if (hover !== null) {
                view.dispatch({ effects: setHoverEffect.of(null) })
                return true
            }
            const position = view.state.selection.main.from + 1
            getHoverTooltip(view, position).then(
                value => view.dispatch({ effects: setHoverEffect.of(value) }),
                () => {}
            )
            return true
        },
    },
    {
        key: 'Enter',
        run(view) {
            const position = positionAtCmPosition(view, view.state.selection.main.from)
            const atEvent = occurrenceAtPosition(view.state, position)
            if (!atEvent) {
                return false
            }
            const { occurrence } = atEvent
            const cmLine = view.state.doc.line(occurrence.range.start.line + 1)
            const cmPos = cmLine.from + occurrence.range.end.character
            const spinner = new Spinner(view, cmPos)
            goToDefinitionAtOccurrence(view, position, occurrence)
                .then(
                    action => action(),
                    () => {}
                )
                .finally(() => spinner.stop())
            return true
        },
    },
    {
        key: 'Mod-ArrowRight',
        run(view) {
            view.state.facet(historyFacet).goForward()
            return true
        },
    },
    {
        key: 'Mod-ArrowLeft',
        run(view) {
            view.state.facet(historyFacet).goBack()
            return true
        },
    },
    {
        key: 'ArrowLeft',
        run(view) {
            view.dispatch({ effects: setHoverEffect.of(null) })
            const position = positionAtCmPosition(view, view.state.selection.main.from)
            const table = view.state.facet(syntaxHighlight)
            const line = position.line
            const occurrence = closestOccurrence(line, table, position, occurrence =>
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
            view.dispatch({ effects: setHoverEffect.of(null) })
            const position = positionAtCmPosition(view, view.state.selection.main.from)
            const table = view.state.facet(syntaxHighlight)
            const line = position.line
            const occurrence = closestOccurrence(line, table, position, occurrence =>
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
            view.dispatch({ effects: setHoverEffect.of(null) })
            const position = positionAtCmPosition(view, view.state.selection.main.from)
            const table = view.state.facet(syntaxHighlight)
            for (let line = position.line + 1; line < table.lineIndex.length; line++) {
                const occurrence = closestOccurrence(line, table, position)
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
            view.dispatch({ effects: setHoverEffect.of(null) })
            const position = positionAtCmPosition(view, view.state.selection.main.from)
            const table = view.state.facet(syntaxHighlight)
            for (let line = position.line - 1; line >= 0; line--) {
                const occurrence = closestOccurrence(line, table, position)
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
            view.dispatch({ effects: setHoverEffect.of(null) })
            view.contentDOM.blur()
            return true
        },
    },
]
export const tokenSelectionKeyBindings = keybindings.flatMap(keybinding => {
    if (keybinding.key === 'ArrowUp') {
        return [keybinding, { ...keybinding, key: 'k' }]
    }
    if (keybinding.key === 'ArrowDown') {
        return [keybinding, { ...keybinding, key: 'j' }]
    }
    if (keybinding.key === 'ArrowLeft') {
        return [keybinding, { ...keybinding, key: 'h' }, { ...keybinding, key: 'Shift-Tab' }]
    }
    if (keybinding.key === 'ArrowRight') {
        return [keybinding, { ...keybinding, key: 'l' }, { ...keybinding, key: 'Tab' }]
    }
    return [keybinding]
})
