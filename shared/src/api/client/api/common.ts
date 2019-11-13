import { ProxyResult, proxyValueSymbol } from '@sourcegraph/comlink'
import { noop } from 'lodash'
import { from, Observable, observable, Subscription } from 'rxjs'
import { mergeMap } from 'rxjs/operators'
import { Subscribable } from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'
import { syncSubscription } from '../../util'
import { asError } from '../../../util/errors'

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
            proxySubscribable =>
                // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                ({
                    // Needed for Rx type check
                    [observable](): Subscribable<T> {
                        return this
                    },
                    subscribe(...args: any[]): Subscription {
                        // Always subscribe with an object because the other side
                        // is unable to tell if a Proxy is a function or an observer object
                        // (they always appear as functions)
                        let proxyObserver: Parameters<typeof proxySubscribable['subscribe']>[0]
                        if (typeof args[0] === 'function') {
                            proxyObserver = {
                                [proxyValueSymbol]: true,
                                next: args[0] || noop,
                                error: args[1] ? err => args[1](asError(err)) : noop,
                                complete: args[2] || noop,
                            }
                        } else {
                            const partialObserver = args[0] || {}
                            proxyObserver = {
                                [proxyValueSymbol]: true,
                                next: partialObserver.next ? val => partialObserver.next(val) : noop,
                                error: partialObserver.error ? err => partialObserver.error(asError(err)) : noop,
                                complete: partialObserver.complete ? () => partialObserver.complete() : noop,
                            }
                        }
                        return syncSubscription(proxySubscribable.subscribe(proxyObserver))
                    },
                } as Subscribable<T>)
        )
    )
