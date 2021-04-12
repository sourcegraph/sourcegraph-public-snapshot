import { Remote, ProxyMarked, proxy, proxyMarker, UnproxyOrClone } from 'comlink'
import { identity } from 'lodash'
import { from, isObservable, Observable, Observer, of } from 'rxjs'
import { map } from 'rxjs/operators'
import { ProviderResult, Subscribable, Unsubscribable } from 'sourcegraph'

import { isAsyncIterable, isPromiseLike, isSubscribable, observableFromAsyncIterable } from '../../util'

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
export const proxySubscribable = <T>(subscribable: Subscribable<T>): ProxySubscribable<T> => ({
    [proxyMarker]: true,
    subscribe(observer): Unsubscribable & ProxyMarked {
        return proxy(
            // Don't pass the proxy to Rx directly because it will try to
            // access Symbol properties that cannot be proxied
            subscribable.subscribe({
                next: value => {
                    // eslint-disable-next-line @typescript-eslint/no-floating-promises
                    observer.next(value as UnproxyOrClone<T>)
                },
                error: error => {
                    // Only pass a few well-known Error properties
                    // TODO should pass all properties serialized recursively, best handled on comlink level
                    // eslint-disable-next-line @typescript-eslint/no-floating-promises
                    observer.error(error && { message: error.message, name: error.name, stack: error.stack })
                },
                complete: () => {
                    // eslint-disable-next-line @typescript-eslint/no-floating-promises
                    observer.complete()
                },
            })
        )
    },
})

export function providerResultToObservable<T, R = T>(
    result: ProviderResult<T>,
    mapFunc: (value: T | undefined | null) => R = identity
): Observable<R> {
    let observable: Observable<R>
    if (result && (isPromiseLike(result) || isObservable<T>(result) || isSubscribable(result))) {
        observable = from(result).pipe(map(mapFunc))
    } else if (isAsyncIterable(result)) {
        observable = observableFromAsyncIterable(result).pipe(map(mapFunc))
    } else {
        observable = of(mapFunc(result))
    }
    return observable
}

interface Deferred<Value = unknown> extends Promise<Value> {
    resolve(value?: Value): void;
    reject(error: Error): void;
}

const createAsyncMutex = <Value,>(): Deferred<Value> => {
    const methodStore = {} as Deferred<Value>;
    const promise = new Promise((resolve, reject) => {
        methodStore.resolve = resolve;
        methodStore.reject = resolve;
    })

    return {...promise, ...methodStore};
}

export async function* observableToAsyncGenerator<D>(observable: Observable<D>): AsyncIterableIterator<D> {
    let asyncMutex = createAsyncMutex<D>();
    let completed = false;

    const subscription = observable.subscribe({
        next: value => {
            const currentAsyncMutex = asyncMutex;
            asyncMutex = createAsyncMutex<D>();

            currentAsyncMutex.resolve(value);
        },
        error: error => {
            const currentAsyncMutex = asyncMutex;
            asyncMutex = createAsyncMutex<D>();
            currentAsyncMutex.reject(error instanceof Error ? error : new Error(String(error)));
        },
        complete: () => {
            completed = true;
            asyncMutex.resolve();
        },
    })

    try {
        while(true) {
            const nextValue = await asyncMutex;

            if (completed) {
                break
            }

            yield nextValue;
        }
    } finally {
        subscription.unsubscribe()
    }
}

