import type { Observable } from 'rxjs'
import { shareReplay } from 'rxjs/operators'
import { type Readable, type Writable, writable, get, type Unsubscriber } from 'svelte/store'

export type LoadingData<D, E = Error> =
    | { loading: true }
    | { loading: false; data: D; error: null }
    | { loading: false; data: null; error: E }

/**
 * Converts a promise to a readable store which emits data, loading and error states.
 * Sometimes load functions return deferred promises and the data needs to be
 * "post processed" in code (i.e. not using {#await}).
 * Usually when working with async data one has to be careful with outdated data.
 * If the load function has been called again we don't want to process the
 * previous data anymore.
 * Using a (reactive) store makes that simpler since Svelte will automatically unsubscribe
 * when the store changes.
 */
export function asStore<T, E = Error>(
    promise: Promise<T>
): Readable<LoadingData<T, E>> & { set(promise: Promise<T>): void } {
    const { subscribe, set } = writable<LoadingData<T, E>>({ loading: true })

    function process(currentPromise: Promise<T>) {
        promise = currentPromise
        currentPromise.then(
            result => {
                if (currentPromise === promise) {
                    set({ loading: false, data: result, error: null })
                }
            },
            error => {
                if (currentPromise === promise) {
                    set({ loading: false, data: null, error })
                }
            }
        )
    }

    process(promise)

    return {
        subscribe,
        set: process,
    }
}

/**
 * Helper function to convert an Observable to a Svelte Readable. Useful when a
 * real Readable is needed to satisfy an interface.
 */
export function readableObservable<T>(observable: Observable<T>): Readable<T> {
    const sharedObservable = observable.pipe(shareReplay(1))
    return {
        subscribe(subscriber) {
            const subscription = sharedObservable.subscribe(subscriber)
            return () => subscription.unsubscribe()
        },
    }
}

/**
 * Returns a helper store that syncs with the currently set store.
 */
export function createForwardStore<T>(store: Writable<T>): Writable<T> & { updateStore(store: Writable<T>): void } {
    const { subscribe, set } = writable<T>(get(store), () => link(store))

    let unsubscribe: Unsubscriber | null = null
    function link(store: Writable<T>): Unsubscriber {
        unsubscribe?.()
        return (unsubscribe = store.subscribe(set))
    }

    return {
        subscribe,
        set(value) {
            store.set(value)
        },
        update(value) {
            store.update(value)
        },
        updateStore(newStore) {
            if (newStore !== store) {
                store = newStore
                link(store)
            }
        },
    }
}
