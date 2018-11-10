import { Unsubscribable } from 'sourcegraph';
/**
 * Manages a map of subscriptions keyed on a numeric ID.
 *
 * @internal
 */
export declare class SubscriptionMap {
    private map;
    /**
     * Adds a new subscription with the given {@link id}.
     *
     * @param id - A unique identifier for this subscription among all other entries in this map.
     * @param subscription - The subscription, unsubscribed when {@link SubscriptionMap#remove} is called.
     * @throws If there already exists an entry with the given {@link id}.
     */
    add(id: number, subscription: Unsubscribable): void;
    /**
     * Unsubscribes the subscription that was previously added with the given {@link id}, and removes it from the
     * map.
     */
    remove(id: number): void;
    /**
     * Unsubscribes all subscriptions in this map and clears it.
     */
    unsubscribe(): void;
}
