import { Unsubscribable } from 'sourcegraph';
/**
 * Manages a set of providers and associates a unique ID with each.
 *
 * @template B - The base provider type.
 * @internal
 */
export declare class ProviderMap<B> {
    private unsubscribeProvider;
    private idSequence;
    private map;
    /**
     * @param unsubscribeProvider - Callback to unsubscribe a provider.
     */
    constructor(unsubscribeProvider: (id: number) => void);
    /**
     * Adds a new provider.
     *
     * @param provider - The provider to add.
     * @returns A newly allocated ID for the provider, unique among all other IDs in this map, and an
     *          unsubscribable for the provider.
     * @throws If there already exists an entry with the given {@link id}.
     */
    add(provider: B): {
        id: number;
        subscription: Unsubscribable;
    };
    /**
     * Returns the provider with the given {@link id}.
     *
     * @template P - The specific provider type for the provider with this {@link id}.
     * @throws If there is no entry with the given {@link id}.
     */
    get<P extends B>(id: number): P;
    /**
     * Unsubscribes the subscription that was previously assigned the given {@link id}, and removes it from the
     * map.
     */
    remove(id: number): void;
    /**
     * Unsubscribes all subscriptions in this map and clears it.
     */
    unsubscribe(): void;
}
