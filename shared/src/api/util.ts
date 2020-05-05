import {
    ProxyMarked,
    transferHandlers,
    ProxyMethods,
    createEndpoint,
    releaseProxy,
    TransferHandler,
    Remote,
} from 'comlink'
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
    const urlTransferHandler: TransferHandler<URL, string> = {
        canHandle: isURL,
        serialize: url => [url.href, []],
        deserialize: urlString => new URL(urlString),
    }
    transferHandlers.set('URL', urlTransferHandler)
}

/**
 * Creates a synchronous Subscription that will unsubscribe the given proxied Subscription asynchronously.
 *
 * @param subscriptionPromise A Promise for a Subscription proxied from the other thread
 */
export const syncSubscription = (subscriptionPromise: Promise<Remote<Unsubscribable & ProxyMarked>>): Subscription =>
    // We cannot pass the proxy subscription directly to Rx because it is a Proxy that looks like a function
    // eslint-disable-next-line @typescript-eslint/no-misused-promises
    new Subscription(async function (this: any) {
        const subscriptionProxy = await subscriptionPromise
        await subscriptionProxy.unsubscribe()
        subscriptionProxy[releaseProxy]()

        this._unsubscribe = null // Workaround: rxjs doesn't null out the reference to this callback
        ;(subscriptionPromise as any) = null
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
