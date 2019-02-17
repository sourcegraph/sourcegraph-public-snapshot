import { ProxyResult, proxyValue } from 'comlink'
import { from, Observable, observable, Subscription } from 'rxjs'
import { mergeMap } from 'rxjs/operators'
import { Subscribable } from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'

/**
 * When a Subscribable is returned from the other thread (wrapped with `proxySubscribable()`),
 * this thread gets a `Promise` for a `Subscribable` _proxy_ where `subscribe()` returns a `Promise<Unsubscribable>`.
 * This function wraps that proxy in a real Rx Observable where `subscribe()` returns an `Unsubscribable` directly as expected.
 *
 * @param proxyPromise The proxy to the `ProxyObservable` in the other thread
 */
export const wrapRemoteObservable = <T>(proxyPromise: Promise<ProxyResult<ProxySubscribable<T>>>): Observable<T> =>
    from(proxyPromise).pipe(
        mergeMap(
            proxy =>
                // tslint:disable-next-line: no-object-literal-type-assertion
                ({
                    // Needed for Rx type check
                    [observable](): Subscribable<T> {
                        return this
                    },
                    subscribe(...args: any): Subscription {
                        const subscription = new Subscription()
                        // tslint:disable-next-line: no-floating-promises
                        proxy.subscribe(...proxyValue(args)).then(s => {
                            subscription.add(s)
                        })
                        return subscription
                    },
                } as Subscribable<T>)
        )
    )
