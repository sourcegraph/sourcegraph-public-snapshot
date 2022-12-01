import { Facet, StateEffect, StateField } from '@codemirror/state'
import { EditorView, PluginValue, ViewPlugin } from '@codemirror/view'

import { isMacPlatform } from '@sourcegraph/common'

const setModifierHeld = StateEffect.define<boolean>()

// State field that is true when the modifier key is held down (Cmd on macOS and
// Ctrl on Linux/Windows).
export const isModifierKeyHeld = StateField.define<boolean>({
    create: () => false,
    update: (value, transactions) => {
        for (const effect of transactions.effects) {
            if (effect.is(setModifierHeld)) {
                value = effect.value
            }
        }
        return value
    },
    provide: field => [
        EditorView.contentAttributes.compute([field], state => ({
            class: state.field(field) ? 'cm-token-selection-clickable' : '',
        })),
    ],
})

// View plugin that makes the cursor look like a pointer when holding down the
// modifier key.
const cmdPointerCursor = ViewPlugin.fromClass(
    class implements PluginValue {
        constructor(public view: EditorView) {
            // Register the lister on document.body so that the cursor looks
            // like pointer even when the CodeMirror content dom is blurred.
            document.body.addEventListener('keydown', this.onKeyDown)
            document.body.addEventListener('keyup', this.onKeyUp)
        }
        public destroy(): void {
            document.body.removeEventListener('keydown', this.onKeyDown)
            document.body.removeEventListener('keyup', this.onKeyUp)
        }
        public onKeyUp = (): void => {
            if (this.view.state.field(isModifierKeyHeld)) {
                this.view.dispatch({ effects: setModifierHeld.of(false) })
            }
        }
        public onKeyDown = (event: KeyboardEvent): void => {
            if (isModifierKey(event) && !this.view.state.field(isModifierKeyHeld)) {
                this.view.dispatch({ effects: setModifierHeld.of(true) })
            }
        }
    }
)

export function isModifierKey(event: KeyboardEvent | MouseEvent): boolean {
    if (isMacPlatform()) {
        return event.metaKey
    }
    return event.ctrlKey
}

export const modifierClickFacet = Facet.define<boolean, boolean>({
    combine: sources => sources[0],
    enables: [isModifierKeyHeld, cmdPointerCursor],
})

export const modifierClickDescription = isMacPlatform() ? 'cmd+click' : 'ctrl+click'
