import { ProxyResult, ProxyValue, proxyValue, proxyValueSymbol } from 'comlink'
import { from, isObservable, Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'
import {
    CompletionObserver,
    ErrorObserver,
    NextObserver,
    PartialObserver,
    ProviderResult,
    Unsubscribable,
} from 'sourcegraph'
import { isPromise, isSubscribable } from '../../util'

type SubscribeArgs<T> =
    | [PartialObserver<T> | undefined]
    | [
          ((value: T) => void) | undefined | null,
          ((error: any) => void) | undefined | null,
          (() => void) | undefined | null
      ]

/**
 * An alternative definition of Subscribable that uses tuples and rest args instead of overloads,
 * because overloads are not persisted through comlink's mapped and conditional types.
 *
 * See https://github.com/Microsoft/TypeScript/issues/29732
 */
export interface SubscribableNoOverloads<T> {
    subscribe(...observer: SubscribeArgs<T>): Unsubscribable
}

/**
 * A Subscribable that can be exposed by comlink to the other thread.
 */
export interface ProxySubscribable<T> extends ProxyValue {
    subscribe(
        ...observer:
            | [

                      | ProxyResult<NextObserver<T> & ProxyValue>
                      | ProxyResult<ErrorObserver<T> & ProxyValue>
                      | ProxyResult<CompletionObserver<T> & ProxyValue>
                      | undefined
              ]
            | [
                  (ProxyResult<((value: T) => void) & ProxyValue> | undefined | null),
                  (ProxyResult<((error: any) => void) & ProxyValue> | undefined | null),
                  (ProxyResult<(() => void) & ProxyValue> | undefined | null)
              ]
    ): Unsubscribable
}

/**
 * Wraps a given Subscribable so that it is exposed by comlink to the other thread.
 *
 * @param subscribable A normal Subscribable (from this thread)
 */
export const proxySubscribable = <T>(subscribable: SubscribableNoOverloads<T>): ProxySubscribable<T> => ({
    [proxyValueSymbol]: true,
    subscribe(...args): Unsubscribable & ProxyValue {
        // cast is needed because of Observer.closed
        return proxyValue(subscribable.subscribe(...(args as SubscribeArgs<T>)))
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
    return proxySubscribable(observable)
}
