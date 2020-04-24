import {
    RemoteObject,
    ProxyMarked,
    transferHandlers,
    ProxyMethods,
    createEndpoint,
    releaseProxy,
} from '@sourcegraph/comlink'
import { Subscription } from 'rxjs'
import { Subscribable, Unsubscribable } from 'sourcegraph'
import { hasProperty } from '../util/types'
import { noop } from 'lodash'

/**
 * Tests whether a value is a WHATWG URL object.
 */
export const isURL = (value: unknown): value is URL =>
    typeof value === 'object' &&
    value !== null &&
    hasProperty('href')(value) &&
    hasProperty('toString')(value) &&
    typeof value.toString === 'function' &&
    // eslint-disable-next-line @typescript-eslint/no-base-to-string
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
        // eslint-disable-next-line @typescript-eslint/no-unsafe-return, @typescript-eslint/no-unsafe-member-access
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
    subscriptionPromise: Promise<RemoteObject<Unsubscribable & ProxyMarked>>
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
export const isPromiseLike = (value: unknown): value is PromiseLike<unknown> =>
    typeof value === 'object' && value !== null && hasProperty('then')(value) && typeof value.then === 'function'

/**
 * Reports whether value is a {@link sourcegraph.Subscribable}.
 */
export const isSubscribable = (value: unknown): value is Subscribable<unknown> =>
    typeof value === 'object' &&
    value !== null &&
    hasProperty('subscribe')(value) &&
    typeof value.subscribe === 'function'

export const addProxyMethods = <T>(value: T): T & ProxyMethods =>
    Object.assign(value, {
        [createEndpoint]: () => Promise.resolve(new MessagePort()),
        [releaseProxy]: noop,
    })
