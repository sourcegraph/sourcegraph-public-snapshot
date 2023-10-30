import { StateEffect, StateField } from '@codemirror/state'
import { EditorView, type PluginValue, ViewPlugin } from '@codemirror/view'

import { isModifierKey } from './utils'

const setModifierHeld = StateEffect.define<boolean>()

/**
 * This extension keeps track of whether the modifier key is pressed or
 * not, and underlines ranges that have a definition available
 * (as determined by the .sg-definition-available class)
 */
export const isModifierKeyHeld = StateField.define<boolean>({
    create: () => false,
    update: (value, transactions) => {
        for (const effect of transactions.effects) {
            if (effect.is(setModifierHeld)) {
                return effect.value
            }
        }
        return value
    },
    provide: self => [
        EditorView.contentAttributes.compute([self], state => ({
            class: state.field(self) ? 'sg-token-selection-clickable' : '',
        })),

        // View plugin that makes the cursor look like a pointer when holding down the
        // modifier key.
        ViewPlugin.fromClass(
            class ModPointerCursor implements PluginValue {
                constructor(private view: EditorView) {
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
                    this.view.dispatch({ effects: setModifierHeld.of(false) })
                }
                public onKeyDown = (event: KeyboardEvent): void => {
                    if (isModifierKey(event)) {
                        this.view.dispatch({ effects: setModifierHeld.of(true) })
                    }
                }
            }
        ),

        EditorView.theme({
            '.sg-token-selection-clickable:hover': {
                cursor: 'pointer',
                // This class is set on eligible ranges in definition.ts
                '& .sg-definition-available': {
                    'text-decoration': 'underline',
                },
            },
        }),
    ],
})
