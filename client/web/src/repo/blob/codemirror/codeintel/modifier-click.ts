import { type Extension, StateEffect, StateField } from '@codemirror/state'
import { EditorView, type PluginValue, ViewPlugin, ViewUpdate } from '@codemirror/view'

import { isMacPlatform } from '@sourcegraph/common'

import { getCodeIntelAPI } from './api'
import { getHoverRange } from './hover'

const setModifierHeld = StateEffect.define<boolean>()

// State field that is true when the modifier key is held down (Cmd on macOS and
// Ctrl on Linux/Windows).
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
    ],
})

const setHasDefinition = StateEffect.define<{ range: { from: number; to: number }; hasDefinition: boolean }>()
const hasDefinition = StateField.define<{ range: { from: number; to: number } | null; hasDefinition: boolean }>({
    create() {
        return { range: null, hasDefinition: false }
    },

    update(value, transaction) {
        const hoverRange = getHoverRange(transaction.state)
        if (hoverRange !== getHoverRange(transaction.startState)) {
            return { range: null, hasDefinition: false }
        }
        for (const effect of transaction.effects) {
            if (
                effect.is(setHasDefinition) &&
                effect.value.range.from === hoverRange?.from &&
                effect.value.range.to === hoverRange.to
            ) {
                return effect.value
            }
        }
        return value
    },

    provide(field) {
        return [
            ViewPlugin.fromClass(
                class DefinitionLoader implements PluginValue {
                    constructor(private view: EditorView) {}

                    update(update: ViewUpdate): void {
                        const hoverRange = getHoverRange(update.state)
                        if (hoverRange && hoverRange !== getHoverRange(update.startState)) {
                            getCodeIntelAPI(update.state)
                                .hasDefinitionAt(hoverRange.from, update.state)
                                .then(hasDefinition => {
                                    this.view.dispatch({
                                        effects: setHasDefinition.of({ range: hoverRange, hasDefinition }),
                                    })
                                })
                        }
                    }
                }
            ),
            EditorView.contentAttributes.computeN([field], state => {
                const { range, hasDefinition } = state.field(field)
                if (range && hasDefinition) {
                    return [{ class: 'sg-definition-available' }]
                }
                return []
            }),
        ]
    },
})

export function isModifierKey(event: KeyboardEvent | MouseEvent): boolean {
    if (isMacPlatform()) {
        return event.metaKey
    }
    return event.ctrlKey
}

const theme = EditorView.theme({
    '.sg-token-selection-clickable:hover': {
        cursor: 'pointer',
        '&.sg-definition-available .selection-highlight-hover': {
            'text-decoration': 'underline',
        },
    },
})

export const modifierClickExtension: Extension = [isModifierKeyHeld, hasDefinition, theme]
