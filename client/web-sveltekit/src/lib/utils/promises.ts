import { type Readable, writable, readonly, derived } from 'svelte/store'

interface PromiseStore<D, E = Error> {
    /**
     * True when the promise is pending, false otherwise.
     * Initial value: false
     */
    pending: Readable<boolean>
    /**
     * The current value or null if the current promise was rejected or is pending.
     * Initial value: null
     */
    value: Readable<D | null>
    /**
     * The current error or null if the current promise was resolved or is pending.
     * Initial value: null
     */
    error: Readable<E | null>
    /**
     * The value of the latest settled promise. While a new promise is pending this will contain
     * the value of the previously settled promise (or null if the promise was rejected).
     * Initial value: null
     */
    latestValue: Readable<D | null>
    /**
     * The value of the latest promise. While a new promise is pending this will contain
     * the value of the previously settled promise (or null if the promise was resolved).
     * Initial value: null
     */
    latestError: Readable<E | null>
    /**
     * Sets the passed promise as the current promise and tracks its status.
     * Does nothing if the same promise as the current one is passed. The argument
     * is optional to make it easier to work with optional data coming from loaders.
     */
    set: (promise?: Promise<D> | null) => void
}

/**
 * Returns multiple stores to track the promises state, resolved value and rejection error.
 * The store ensures that `value` is updated with latest resolved promise.
 */
export function createPromiseStore<D, E = Error>(): PromiseStore<Awaited<D>, E> {
    let currentPromise: Promise<Awaited<D>> | null | undefined

    const pending = writable<boolean>(false)
    const value = writable<Awaited<D> | null>(null)
    const error = writable<E | null>(null)

    function resolve(promise?: Promise<Awaited<D>> | null) {
        currentPromise = promise
        if (!promise) {
            value.set(null)
            error.set(null)
            pending.set(false)
            return
        }

        pending.set(true)
        promise.then(
            result => {
                if (currentPromise === promise) {
                    value.set(result)
                    error.set(null)
                    pending.set(false)
                }
            },
            error_ => {
                if (currentPromise === promise) {
                    value.set(null)
                    error.set(error_)
                    pending.set(false)
                }
            }
        )
    }

    resolve(currentPromise)

    return {
        pending: readonly(pending),
        value: derived([pending, value], ([$pending, $value]) => ($pending ? null : $value)),
        error: derived([pending, error], ([$pending, $error]) => ($pending ? null : $error)),
        latestValue: readonly(value),
        latestError: readonly(error),
        set: promise => {
            if (promise !== currentPromise) {
                resolve(promise)
            }
        },
    }
}
