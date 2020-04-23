import { Remote, ProxyMarked, proxy, proxyMarker, UnproxyOrClone } from '@sourcegraph/comlink'
import { from, isObservable, Observable, Observer, of } from 'rxjs'
import { map } from 'rxjs/operators'
import { ProviderResult, Subscribable, Unsubscribable } from 'sourcegraph'
import { isPromiseLike, isSubscribable } from '../../util'

/**
 * A Subscribable that can be exposed by comlink to the other thread.
 * Only allows full object Observers to avoid complex type checking against proxies.
 */
export interface ProxySubscribable<T> extends ProxyMarked {
    subscribe(observer: Remote<Observer<T> & ProxyMarked>): Unsubscribable & ProxyMarked
}

/**
 * Wraps a given Subscribable so that it is exposed by comlink to the other thread.
 *
 * @param subscribable A normal Subscribable (from this thread)
 */
const proxySubscribable = <T>(subscribable: Subscribable<T>): ProxySubscribable<T> => ({
    [proxyMarker]: true,
    subscribe(observer): Unsubscribable & ProxyMarked {
        return proxy(
            // Don't pass the proxy to Rx directly because it will try to
            // access Symbol properties that cannot be proxied
            subscribable.subscribe({
                next: val => {
                    // eslint-disable-next-line @typescript-eslint/no-floating-promises
                    observer.next(val as UnproxyOrClone<T>)
                },
                error: err => {
                    // Only pass a few well-known Error properties
                    // TODO should pass all properties serialized recursively, best handled on comlink level
                    // eslint-disable-next-line @typescript-eslint/no-floating-promises
                    observer.error(err && { message: err.message, name: err.name, stack: err.stack })
                },
                complete: () => {
                    // eslint-disable-next-line @typescript-eslint/no-floating-promises
                    observer.complete()
                },
            })
        )
    },
})

/**
 * Returns a Subscribable that can be proxied by comlink.
 *
 * @param result The result returned by the provider
 * @param mapFunc A function to map the result into a value to be transmitted to the other thread
 */
export function toProxyableSubscribable<T, R>(
    result: ProviderResult<T>,
    mapFunc: (value: T | undefined | null) => R
): ProxySubscribable<R> {
    let observable: Observable<R>
    if (result && (isPromiseLike(result) || isObservable<T>(result) || isSubscribable(result))) {
        observable = from(result).pipe(map(mapFunc))
    } else {
        observable = of(mapFunc(result))
    }
    return proxySubscribable(observable)
}
