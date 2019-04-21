import { proxy, proxyMarker, ProxyResult, ProxyValue, UnproxyOrClone } from '@sourcegraph/comlink'
import { from, isObservable, Observable, Observer, of } from 'rxjs'
import { map } from 'rxjs/operators'
import { ProviderResult, Subscribable, Unsubscribable } from 'sourcegraph'
import { isPromise, isSubscribable } from '../../util'

/**
 * A Subscribable that can be exposed by comlink to the other thread.
 * Only allows full object Observers to avoid complex type checking against proxies.
 */
export interface ProxySubscribable<T> extends ProxyValue {
    subscribe(observer: ProxyResult<Observer<T> & ProxyValue>): Unsubscribable & ProxyValue
}

/**
 * Wraps a given Subscribable so that it is exposed by comlink to the other thread.
 *
 * @param subscribable A normal Subscribable (from this thread)
 */
export const proxySubscribable = <T>(subscribable: Subscribable<T>): ProxySubscribable<T> => ({
    [proxyMarker]: true,
    subscribe(observer): Unsubscribable & ProxyValue {
        return proxy(
            // Don't pass the proxy to Rx directly because it will try to
            // access Symbol properties that cannot be proxied
            subscribable.subscribe({
                next: val => {
                    // tslint:disable-next-line: no-floating-promises
                    observer.next(val as UnproxyOrClone<T>)
                },
                error: err => {
                    // Only pass a few well-known Error properties
                    // TODO should pass all properties serialized recursively, best handled on comlink level
                    // tslint:disable-next-line: no-floating-promises
                    observer.error(err && { message: err.message, name: err.name, code: err.code, stack: err.stack })
                },
                complete: () => {
                    // tslint:disable-next-line: no-floating-promises
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
    if (result && (isPromise(result) || isObservable<T>(result) || isSubscribable(result))) {
        observable = from(result).pipe(map(mapFunc))
    } else {
        observable = of(mapFunc(result))
    }
    return proxySubscribable(
        new Observable(sub => {
            const u = observable.subscribe(sub)
            // console.count('SUBSCRIBE!+1')
            return () => {
                u.unsubscribe()
                // console.count('UNSUBSCRIBE!-1')
            }
        })
    )
}
