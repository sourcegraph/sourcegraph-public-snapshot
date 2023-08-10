import { type Observable, Subject } from 'rxjs'

/**
 * Simple state container that can be read synchronously and subscribed to to be notified of changes
 */
export interface ObservableStateContainer<S extends object> {
    /**
     * The current state
     */
    readonly values: Readonly<S>

    /**
     * Emits after `values` was updated with the new state.
     */
    readonly updates: Observable<Readonly<S>>

    /**
     * Update `values` and emit the new state on `updates`.
     * Similar to React's `setState()`, but synchronous and without callbacks.
     * It is therefor safe to derive the new state from the current state,
     * and read the new state right after updating.
     */
    update<K extends keyof S>(state: Pick<S, K> | S): void
}

export const createObservableStateContainer = <S extends object>(initial: S): ObservableStateContainer<S> => {
    const container = {
        values: initial,
        updates: new Subject<S>(),

        update<K extends keyof S>(state: Pick<S, K> | S): void {
            this.values = { ...(this.values as object), ...(state as object) } as Readonly<S>
            this.updates.next(this.values)
        },
    }
    return container
}
