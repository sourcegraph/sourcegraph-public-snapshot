import { type Readable, writable, readonly } from 'svelte/store'

/**
 * Successful result of a promise.
 */
interface ResultSuccess<T> {
    value: T
    error: null
    pending: false
}

/**
 * Rejected result of a promise.
 */
interface ResultError<E> {
    value: null
    error: E
    pending: false
}

/**
 * Pending result of a promise. `value` and `error` can contain the
 * latest resolved value or rejection error.
 */
interface ResultPending<T, E> {
    value: T | null
    error: E | null
    pending: true
}

export type Loadable<T, E = Error> = ResultSuccess<T> | ResultError<E> | ResultPending<T, E>

interface PromiseStore<D, E = Error> extends Readable<Loadable<D | null, E>> {
    /**
     * Sets the passed promise as the current promise and tracks its status.
     * Does nothing if the same promise as the current one is passed. The argument
     * is optional to make it easier to work with optional data coming from loaders.
     */
    set: (promise?: PromiseLike<D> | null) => void

    /**
     * Resets the store to its initial state.
     */
    reset: () => void
}

const initialLoadable: Loadable<any, any> = { value: null, error: null, pending: true }

/**
 * Returns multiple stores to track the promises state, resolved value and rejection error.
 * The store ensures that `value` is updated with latest resolved promise.
 */
export function createPromiseStore<D, E = Error>(): PromiseStore<D, E> {
    let currentPromise: PromiseLike<D> | null | undefined

    const resultStore = writable<Loadable<D | null, E>>(initialLoadable)

    function resolve(promise?: PromiseLike<D> | null) {
        currentPromise = promise
        if (!promise) {
            resultStore.set({ value: null, error: null, pending: false })
            return
        }

        resultStore.update($result => ({ ...$result, pending: true }))
        promise.then(
            result => {
                if (currentPromise === promise) {
                    resultStore.update(() => ({ value: result, error: null, pending: false }))
                }
            },
            error => {
                if (currentPromise === promise) {
                    resultStore.update(() => ({ value: null, error: error, pending: false }))
                }
            }
        )
    }

    resolve(currentPromise)

    return {
        ...readonly(resultStore),
        set: promise => {
            if (promise !== currentPromise) {
                resolve(promise)
            }
        },
        reset() {
            resultStore.set(initialLoadable)
        },
    }
}

/**
 * Returns a store that publishes updates when the promise is resolved or rejected.
 */
export function toReadable<D, E = Error>(promise: PromiseLike<D>): Readable<Loadable<D, E>> {
    const { subscribe, set } = writable<Loadable<D, E>>({ value: null, error: null, pending: true })
    promise.then(
        result => {
            set({ value: result, error: null, pending: false })
        },
        error => {
            set({ value: null, error: error, pending: false })
        }
    )
    return { subscribe }
}

// Returns a promise that is guaranteed to take at least `delayMillis` milliseconds to resolve.
// If the wrapped promise resolves before then, the returned promise will wait until `delayMillis`
// has elapsed before resolving.
export async function delay<T>(promise: Promise<T>, delayMillis: number): Promise<T> {
    const [awaited] = await Promise.all([promise, new Promise(resolve => setTimeout(resolve, delayMillis))])
    return awaited
}
