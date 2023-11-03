import { type EditorState, type Extension, StateEffect, StateField } from '@codemirror/state'
import { type EditorView, type PluginValue, ViewPlugin, type ViewUpdate } from '@codemirror/view'
import type { Observable, Subscription } from 'rxjs'

import { isMacPlatform } from '@sourcegraph/common'

/**
 * Represents a value that can be updated by the loader extension.
 * The second type parameter is a bit annoying but seems necessary
 * to properly express the constraints for the {@method update} method.
 * It should be set to the class that implements this interface.
 */
export interface UpdateableValue<T, U> {
    key: unknown
    update(value: T): U
    isPending: boolean
}

interface LoaderExtensionSpec<Response, Input, Value extends UpdateableValue<Response, Value>> {
    input: (state: EditorState) => readonly Input[]
    create: (input: Input) => Value
    load(value: Value, state: EditorState): Observable<Response>
    provide: (field: StateField<Value[]>) => Extension
}

export function createLoaderExtension<Response, Input, Value extends UpdateableValue<Response, Value>>(
    spec: LoaderExtensionSpec<Response, Input, Value>
): Extension {
    const updateValueEffect = StateEffect.define<{ key: unknown; result: Response }>()

    return StateField.define<Value[]>({
        create(state) {
            return spec.input(state).map(input => spec.create(input))
        },

        update(values, transaction) {
            const newValues = syncValues({
                values,
                currentInput: spec.input(transaction.state),
                previousInput: spec.input(transaction.startState),
                create: input => spec.create(input),
                update: value => {
                    for (const effect of transaction.effects) {
                        if (effect.is(updateValueEffect) && effect.value.key === value.key) {
                            return value.update(effect.value.result)
                        }
                    }
                    return value
                },
            })
            return newValues
        },

        provide: field => [
            ViewPlugin.fromClass(
                class Loader implements PluginValue {
                    private loading: Map<unknown, Subscription> = new Map()

                    constructor(private view: EditorView) {
                        for (const value of view.state.field(field)) {
                            if (value.isPending) {
                                this.load(value, view.state)
                            }
                        }
                    }

                    public update(update: ViewUpdate): void {
                        const values = update.state.field(field)
                        if (values !== update.startState.field(field)) {
                            const seen = new Set<unknown>()

                            // Start loading for new values
                            for (const value of values) {
                                seen.add(value.key)
                                if (value.isPending && !this.loading.has(value.key)) {
                                    this.load(value, update.state)
                                }
                            }

                            // Remove subscriptions for values that don't exist anymore
                            for (const [key, subscription] of this.loading) {
                                if (!seen.has(key)) {
                                    subscription.unsubscribe()
                                    this.loading.delete(key)
                                }
                            }
                        }
                    }

                    private load(value: Value, state: EditorState): void {
                        this.loading.set(
                            value.key,
                            spec.load(value, state).subscribe({
                                next: result => {
                                    this.view.dispatch({ effects: updateValueEffect.of({ key: value.key, result }) })
                                },
                                complete: () => {
                                    this.loading.delete(value.key)
                                },
                            })
                        )
                    }
                }
            ),
            spec.provide(field),
        ],
    })
}

/**
 * Helper function for updating/syncing a list of values. {@param create} is
 * called for every element in {@param current} that is not in {@param previous}.
 * {@param update} is called for every element of {@param current} that also
 * exists in {@param previous}. If the final list differs from {@param values} the
 * new list is returned.
 */
export function syncValues<T, U>({
    values,
    previousInput,
    currentInput,
    create,
    update,
}: {
    values: U[]
    previousInput: readonly T[]
    currentInput: readonly T[]
    create: (value: T) => U
    update: (value: U) => U
}): U[] {
    let updated = previousInput.length !== currentInput.length
    const newValues: U[] = new Array(currentInput.length)

    for (let i = 0; i < currentInput.length; i++) {
        const known = previousInput.indexOf(currentInput[i])

        if (known === -1) {
            updated = true
            newValues[i] = create(currentInput[i])
        } else {
            const newValue = (newValues[i] = update(values[known]))
            if (i !== known || newValue !== values[known]) {
                updated = true
            }
        }
    }

    return updated ? newValues : values
}

/**
 * Returns true if the modifier key (Mac: Meta, others: Ctrl) is pressed.
 */
export function isModifierKey(event: KeyboardEvent | MouseEvent): boolean {
    if (isMacPlatform()) {
        return event.metaKey
    }
    return event.ctrlKey
}
