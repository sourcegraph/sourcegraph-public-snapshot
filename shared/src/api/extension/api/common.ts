import { from, isObservable, Observable, of, Subscribable as RxJSSubscribable } from 'rxjs'
import { map } from 'rxjs/operators'
import { Subscribable, Unsubscribable } from 'sourcegraph'
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

/**
 * Returns an Observable that emits the provider result.
 */
export function toProviderResultObservable<T, R>(
    result: Promise<T | undefined | null | Subscribable<T | undefined | null>>,
    mapFunc: (value: T | undefined | null) => R | undefined | null
): Observable<R | undefined | null> {
    return new Observable<R | undefined | null>(observer => {
        result
            .then(result => {
                let observable: Observable<R | undefined | null>
                if (result && (isPromise(result) || isObservable(result) || isSubscribable(result))) {
                    observable = from(result as Promise<any> | RxJSSubscribable<any>).pipe(map(mapFunc))
                } else {
                    observable = of(mapFunc(result))
                }
                observable.subscribe(observer)
            })
            .catch(err => observer.error(err))
    })
}
