import { Remote, proxyMarker, releaseProxy, ProxyMethods } from 'comlink'
import { noop } from 'lodash'
import { from, Observable, observable as symbolObservable, Subscription } from 'rxjs'
import { mergeMap, finalize } from 'rxjs/operators'
import { Subscribable } from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'
import { syncSubscription } from '../../util'
import { asError } from '../../../util/errors'

/**
 * An ordinary Observable linked to an Observable in another thread through a `MessagePort`.
 */
export interface RemoteObservable<T> extends Observable<T>, Pick<ProxyMethods, typeof releaseProxy> {}

/**
 * When a Subscribable is returned from the other thread (wrapped with `proxySubscribable()`),
 * this thread gets a `Promise` for a `Subscribable` _proxy_ where `subscribe()` returns a `Promise<Unsubscribable>`.
 * This function wraps that proxy in a real Rx Observable where `subscribe()` returns a `Subscription` directly as expected.
 *
 * The returned Observable is augmented with the `releaseProxy` method from comlink to release the underlying `MessagePort`.
 *
 * @param proxyPromise The proxy to the `ProxyObservable` in the other thread
 */
export const wrapRemoteObservable = <T>(proxyPromise: Promise<Remote<ProxySubscribable<T>>>): RemoteObservable<T> => {
    const proxyReleaseSubscription = new Subscription()
    const observable = from(proxyPromise).pipe(
        mergeMap(
            (proxySubscribable): Subscribable<T> => {
                proxyReleaseSubscription.add(() => proxySubscribable[releaseProxy]())
                return {
                    // Needed for Rx type check
                    [symbolObservable](): Subscribable<T> {
                        return this
                    },
                    subscribe(...args: any[]): Subscription {
                        // Always subscribe with an object because the other side
                        // is unable to tell if a Proxy is a function or an observer object
                        // (they always appear as functions)
                        let proxyObserver: Parameters<typeof proxySubscribable['subscribe']>[0]
                        if (typeof args[0] === 'function') {
                            proxyObserver = {
                                [proxyMarker]: true,
                                next: args[0] || noop,
                                error: args[1] ? err => args[1](asError(err)) : noop,
                                complete: args[2] || noop,
                            }
                        } else {
                            const partialObserver = args[0] || {}
                            proxyObserver = {
                                [proxyMarker]: true,
                                next: partialObserver.next ? val => partialObserver.next(val) : noop,
                                error: partialObserver.error ? err => partialObserver.error(asError(err)) : noop,
                                complete: partialObserver.complete ? () => partialObserver.complete() : noop,
                            }
                        }
                        return syncSubscription(proxySubscribable.subscribe(proxyObserver))
                    },
                }
            }
        )
    )
    return Object.assign(observable, {
        [releaseProxy]: () => proxyReleaseSubscription.unsubscribe(),
    })
}

/**
 * Releases the underlying MessagePort of a remote Observable when it completes or is unsubscribed from.
 *
 * Important: This will prevent resubscribing to the Observable. Only use this operator in a scope where it is known
 * that no resubscriptions can happen after completion, e.g. in a `switchMap()` callback.
 *
 * Must be used as the first parameter to `pipe()`, because the source must be a `RemoteObservable`.
 */
export const finallyReleaseProxy = <T>() => (
    source: Observable<T> & Partial<Pick<ProxyMethods, typeof releaseProxy>>
) => {
    if (!source[releaseProxy]) {
        console.warn('finallyReleaseProxy() used on Observable with no [releaseProxy] method')
        return source
    }
    return source.pipe(finalize(() => source[releaseProxy]!()))
}
