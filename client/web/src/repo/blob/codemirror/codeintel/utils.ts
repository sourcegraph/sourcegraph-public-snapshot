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
    /**
     * Unique key that identifies this value. This value is used to
     * associate data from the load function with the input value.
     */
    key: unknown

    /**
     * If true the load function should be called for this value.
     */
    isPending: boolean

    /**
     * Called with the data returned from the load function.
     */
    update(value: T): U
}

interface LoaderExtensionSpec<Response, Input, Value extends UpdateableValue<Response, Value>> {
    /**
     * How to get the input from the current editor state. This is also used to
     * determine whether the input changed.
     */
    input: (state: EditorState) => readonly Input[]
    /**
     * This function converts new Input values to another representation. This is useful
     * when additional data needs to be associated with the input or when not every
     * input value should be handled by the load function, as indicated by the value's
     * `isPending` property.
     */
    create: (input: Input) => Value
    /**
     * Called for each Value for which `load` hasn't been called before and whose
     * `isPending` property is `true`. The value's `update` method will be called
     * with the received Response.
     */
    load(value: Value, state: EditorState): Observable<Response>

    /**
     * Returns other extensions that use these values as input.
     */
    provide: (field: StateField<Value[]>) => Extension
}

/**
 * This function/extension implements a common pattern among the other extensions:
 * Receive input, load some data and provide some output (i.e. input to other extensions).
 * What makes this tricky is that the input might change while data is being loaded, cause the
 * outstanding data to be stale. The returned extension handles this case.
 */
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
