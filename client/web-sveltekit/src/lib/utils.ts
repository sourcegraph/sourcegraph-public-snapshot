import type { Observable } from 'rxjs'
import { shareReplay } from 'rxjs/operators'
import { type Readable, writable } from 'svelte/store'

export type LoadingData<D, E> =
    | { loading: true }
    | { loading: false; data: D; error: null }
    | { loading: false; data: null; error: E }

/**
 * Converts a promise to a readable store which emits loading and error data.
 * This is useful in loader functions to prevent SvelteKit from waiting for the
 * promise to resolve before rendering the page.
 */
export function asStore<T, E = Error>(promise: Promise<T>): Readable<LoadingData<T, E>> {
    const { subscribe, set } = writable<LoadingData<T, E>>({ loading: true })
    promise.then(
        result => set({ loading: false, data: result, error: null }),
        error => set({ loading: false, data: null, error })
    )

    return {
        subscribe,
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
