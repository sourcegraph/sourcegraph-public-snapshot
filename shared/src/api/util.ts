import { ProxiedObject, ProxyValue, transferHandlers } from '@sourcegraph/comlink'
import { Subscription } from 'rxjs'
import { Subscribable, Unsubscribable } from 'sourcegraph'

/**
 * Tests whether a value is a WHATWG URL object.
 */
export const isURL = (value: any): value is URL =>
    typeof value !== 'undefined' &&
    value !== null &&
    typeof value.toString === 'function' &&
    value.href === value.toString()

/**
 * Registers global comlink transfer handlers.
 * This needs to be called before using comlink.
 * Idempotent.
 */
export function registerComlinkTransferHandlers(): void {
    transferHandlers.set('URL', {
        canHandle: isURL,
        // TODO the comlink types could be better here to avoid the any
        serialize: (url: any) => url.href,
        deserialize: (urlString: any) => new URL(urlString),
    })
}

/**
 * Creates a synchronous Subscription that will unsubscribe the given proxied Subscription asynchronously.
 *
 * @param subscriptionPromise A Promise for a Subscription proxied from the other thread
 */
export const syncSubscription = (
    subscriptionPromise: Promise<ProxiedObject<Unsubscribable & ProxyValue>>
): Subscription =>
    // We cannot pass the proxy subscription directly to Rx because it is a Proxy that looks like a function
    new Subscription(() => {
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        subscriptionPromise.then(proxySubscription => {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            proxySubscription.unsubscribe()
        })
    })

/**
 * Runs f and returns a resolved promise with its value or a rejected promise with its exception,
 * regardless of whether it returns a promise or not.
 */
export const tryCatchPromise = async <T>(f: () => T | Promise<T>): Promise<T> => f()

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
