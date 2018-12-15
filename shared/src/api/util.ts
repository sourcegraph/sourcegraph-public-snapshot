import { Subscribable } from 'sourcegraph'

/**
 * Runs f and returns a resolved promise with its value or a rejected promise with its exception,
 * regardless of whether it returns a promise or not.
 */
export function tryCatchPromise<T>(f: () => T | Promise<T>): Promise<T> {
    try {
        return Promise.resolve(f())
    } catch (err) {
        return Promise.reject(err)
    }
}

/**
 * Reports whether value is a Promise.
 */
export function isPromise(value: any): value is Promise<any> {
    return typeof value.then === 'function'
}

/**
 * Reports whether value is a {@link sourcegraph.Subscribable}.
 */
export function isSubscribable(value: any): value is Subscribable<any> {
    return typeof value.subscribe === 'function'
}

export interface PromiseCallback<T> {
    resolve: (p: T | Promise<T>) => void
}
