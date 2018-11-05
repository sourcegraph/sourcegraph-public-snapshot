import { Unsubscribable } from 'sourcegraph'

/**
 * Manages a map of subscriptions keyed on a numeric ID.
 *
 * @internal
 */
export class SubscriptionMap {
    private map = new Map<number, Unsubscribable>()

    /**
     * Adds a new subscription with the given {@link id}.
     *
     * @param id - A unique identifier for this subscription among all other entries in this map.
     * @param subscription - The subscription, unsubscribed when {@link SubscriptionMap#remove} is called.
     * @throws If there already exists an entry with the given {@link id}.
     */
    public add(id: number, subscription: Unsubscribable): void {
        if (this.map.has(id)) {
            throw new Error(`subscription already exists with ID ${id}`)
        }
        this.map.set(id, subscription)
    }

    /**
     * Unsubscribes the subscription that was previously added with the given {@link id}, and removes it from the
     * map.
     */
    public remove(id: number): void {
        const subscription = this.map.get(id)
        if (!subscription) {
            throw new Error(`no subscription with ID ${id}`)
        }
        try {
            subscription.unsubscribe()
        } finally {
            this.map.delete(id)
        }
    }

    /**
     * Unsubscribes all subscriptions in this map and clears it.
     */
    public unsubscribe(): void {
        try {
            for (const subscription of this.map.values()) {
                subscription.unsubscribe()
            }
        } finally {
            this.map.clear()
        }
    }
}
