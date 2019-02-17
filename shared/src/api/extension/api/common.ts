import { ProxyResult, ProxyValue, proxyValue, proxyValueSymbol } from 'comlink'
import { from, isObservable, Observable, observable, of, Subscription } from 'rxjs'
import { map, mergeMap } from 'rxjs/operators'
import {
    CompletionObserver,
    ErrorObserver,
    NextObserver,
    PartialObserver,
    ProviderResult,
    Subscribable,
    Unsubscribable,
} from 'sourcegraph'
import { isPromise, isSubscribable } from '../../util'

/**
 * Manages a set of providers and associates a unique ID with each.
 *
 * @template B - The base provider type.
 * @internal
 */
export class ProviderMap<B> {
    private idSequence = 0
    private map = new Map<number, B>()

    /**
     * @param unsubscribeProvider - Callback to unsubscribe a provider.
     */
    constructor(private unsubscribeProvider: (id: number) => void) {}

    /**
     * Adds a new provider.
     *
     * @param provider - The provider to add.
     * @returns A newly allocated ID for the provider, unique among all other IDs in this map, and an
     *          unsubscribable for the provider.
     * @throws If there already exists an entry with the given {@link id}.
     */
    public add(provider: B): { id: number; subscription: Unsubscribable } {
        const id = this.idSequence
        this.map.set(id, provider)
        this.idSequence++
        return { id, subscription: { unsubscribe: () => this.remove(id) } }
    }

    /**
     * Returns the provider with the given {@link id}.
     *
     * @template P - The specific provider type for the provider with this {@link id}.
     * @throws If there is no entry with the given {@link id}.
     */
    public get<P extends B>(id: number): P {
        const provider = this.map.get(id) as P
        if (provider === undefined) {
            throw new Error(`no provider with ID ${id}`)
        }
        return provider
    }

    /**
     * Unsubscribes the subscription that was previously assigned the given {@link id}, and removes it from the
     * map.
     */
    public remove(id: number): void {
        if (!this.map.has(id)) {
            throw new Error(`no provider with ID ${id}`)
        }
        try {
            this.unsubscribeProvider(id)
        } finally {
            this.map.delete(id)
        }
    }

    /**
     * Unsubscribes all subscriptions in this map and clears it.
     */
    public unsubscribe(): void {
        try {
            for (const id of this.map.keys()) {
                this.unsubscribeProvider(id)
            }
        } finally {
            this.map.clear()
        }
    }
}

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

/**
 * When a Subscribable is returned from the other thread (wrapped with `proxySubscribable()`),
 * this thread gets a `Promise` for a `Subscribable` _proxy_ where `subscribe()` returns a `Promise<Unsubscribable>`.
 * This function wraps that proxy in a real Rx Observable where `subscribe()` returns an `Unsubscribable` directly as expected.
 *
 * @param proxyPromise
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
