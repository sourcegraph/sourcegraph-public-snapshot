/**
 * Manages a set of providers and associates a unique ID with each.
 *
 * @template B - The base provider type.
 * @internal
 */
export class ProviderMap {
    /**
     * @param unsubscribeProvider - Callback to unsubscribe a provider.
     */
    constructor(unsubscribeProvider) {
        this.unsubscribeProvider = unsubscribeProvider;
        this.idSequence = 0;
        this.map = new Map();
    }
    /**
     * Adds a new provider.
     *
     * @param provider - The provider to add.
     * @returns A newly allocated ID for the provider, unique among all other IDs in this map, and an
     *          unsubscribable for the provider.
     * @throws If there already exists an entry with the given {@link id}.
     */
    add(provider) {
        const id = this.idSequence;
        this.map.set(id, provider);
        this.idSequence++;
        return { id, subscription: { unsubscribe: () => this.remove(id) } };
    }
    /**
     * Returns the provider with the given {@link id}.
     *
     * @template P - The specific provider type for the provider with this {@link id}.
     * @throws If there is no entry with the given {@link id}.
     */
    get(id) {
        const provider = this.map.get(id);
        if (provider === undefined) {
            throw new Error(`no provider with ID ${id}`);
        }
        return provider;
    }
    /**
     * Unsubscribes the subscription that was previously assigned the given {@link id}, and removes it from the
     * map.
     */
    remove(id) {
        if (!this.map.has(id)) {
            throw new Error(`no provider with ID ${id}`);
        }
        try {
            this.unsubscribeProvider(id);
        }
        finally {
            this.map.delete(id);
        }
    }
    /**
     * Unsubscribes all subscriptions in this map and clears it.
     */
    unsubscribe() {
        try {
            for (const id of this.map.keys()) {
                this.unsubscribeProvider(id);
            }
        }
        finally {
            this.map.clear();
        }
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiY29tbW9uLmpzIiwic291cmNlUm9vdCI6InNyYy8iLCJzb3VyY2VzIjpbImV4dGVuc2lvbi9hcGkvY29tbW9uLnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUVBOzs7OztHQUtHO0FBQ0gsTUFBTSxPQUFPLFdBQVc7SUFJcEI7O09BRUc7SUFDSCxZQUFvQixtQkFBeUM7UUFBekMsd0JBQW1CLEdBQW5CLG1CQUFtQixDQUFzQjtRQU5yRCxlQUFVLEdBQUcsQ0FBQyxDQUFBO1FBQ2QsUUFBRyxHQUFHLElBQUksR0FBRyxFQUFhLENBQUE7SUFLOEIsQ0FBQztJQUVqRTs7Ozs7OztPQU9HO0lBQ0ksR0FBRyxDQUFDLFFBQVc7UUFDbEIsTUFBTSxFQUFFLEdBQUcsSUFBSSxDQUFDLFVBQVUsQ0FBQTtRQUMxQixJQUFJLENBQUMsR0FBRyxDQUFDLEdBQUcsQ0FBQyxFQUFFLEVBQUUsUUFBUSxDQUFDLENBQUE7UUFDMUIsSUFBSSxDQUFDLFVBQVUsRUFBRSxDQUFBO1FBQ2pCLE9BQU8sRUFBRSxFQUFFLEVBQUUsWUFBWSxFQUFFLEVBQUUsV0FBVyxFQUFFLEdBQUcsRUFBRSxDQUFDLElBQUksQ0FBQyxNQUFNLENBQUMsRUFBRSxDQUFDLEVBQUUsRUFBRSxDQUFBO0lBQ3ZFLENBQUM7SUFFRDs7Ozs7T0FLRztJQUNJLEdBQUcsQ0FBYyxFQUFVO1FBQzlCLE1BQU0sUUFBUSxHQUFHLElBQUksQ0FBQyxHQUFHLENBQUMsR0FBRyxDQUFDLEVBQUUsQ0FBTSxDQUFBO1FBQ3RDLElBQUksUUFBUSxLQUFLLFNBQVMsRUFBRTtZQUN4QixNQUFNLElBQUksS0FBSyxDQUFDLHVCQUF1QixFQUFFLEVBQUUsQ0FBQyxDQUFBO1NBQy9DO1FBQ0QsT0FBTyxRQUFRLENBQUE7SUFDbkIsQ0FBQztJQUVEOzs7T0FHRztJQUNJLE1BQU0sQ0FBQyxFQUFVO1FBQ3BCLElBQUksQ0FBQyxJQUFJLENBQUMsR0FBRyxDQUFDLEdBQUcsQ0FBQyxFQUFFLENBQUMsRUFBRTtZQUNuQixNQUFNLElBQUksS0FBSyxDQUFDLHVCQUF1QixFQUFFLEVBQUUsQ0FBQyxDQUFBO1NBQy9DO1FBQ0QsSUFBSTtZQUNBLElBQUksQ0FBQyxtQkFBbUIsQ0FBQyxFQUFFLENBQUMsQ0FBQTtTQUMvQjtnQkFBUztZQUNOLElBQUksQ0FBQyxHQUFHLENBQUMsTUFBTSxDQUFDLEVBQUUsQ0FBQyxDQUFBO1NBQ3RCO0lBQ0wsQ0FBQztJQUVEOztPQUVHO0lBQ0ksV0FBVztRQUNkLElBQUk7WUFDQSxLQUFLLE1BQU0sRUFBRSxJQUFJLElBQUksQ0FBQyxHQUFHLENBQUMsSUFBSSxFQUFFLEVBQUU7Z0JBQzlCLElBQUksQ0FBQyxtQkFBbUIsQ0FBQyxFQUFFLENBQUMsQ0FBQTthQUMvQjtTQUNKO2dCQUFTO1lBQ04sSUFBSSxDQUFDLEdBQUcsQ0FBQyxLQUFLLEVBQUUsQ0FBQTtTQUNuQjtJQUNMLENBQUM7Q0FDSiIsInNvdXJjZXNDb250ZW50IjpbImltcG9ydCB7IFVuc3Vic2NyaWJhYmxlIH0gZnJvbSAnc291cmNlZ3JhcGgnXG5cbi8qKlxuICogTWFuYWdlcyBhIHNldCBvZiBwcm92aWRlcnMgYW5kIGFzc29jaWF0ZXMgYSB1bmlxdWUgSUQgd2l0aCBlYWNoLlxuICpcbiAqIEB0ZW1wbGF0ZSBCIC0gVGhlIGJhc2UgcHJvdmlkZXIgdHlwZS5cbiAqIEBpbnRlcm5hbFxuICovXG5leHBvcnQgY2xhc3MgUHJvdmlkZXJNYXA8Qj4ge1xuICAgIHByaXZhdGUgaWRTZXF1ZW5jZSA9IDBcbiAgICBwcml2YXRlIG1hcCA9IG5ldyBNYXA8bnVtYmVyLCBCPigpXG5cbiAgICAvKipcbiAgICAgKiBAcGFyYW0gdW5zdWJzY3JpYmVQcm92aWRlciAtIENhbGxiYWNrIHRvIHVuc3Vic2NyaWJlIGEgcHJvdmlkZXIuXG4gICAgICovXG4gICAgY29uc3RydWN0b3IocHJpdmF0ZSB1bnN1YnNjcmliZVByb3ZpZGVyOiAoaWQ6IG51bWJlcikgPT4gdm9pZCkge31cblxuICAgIC8qKlxuICAgICAqIEFkZHMgYSBuZXcgcHJvdmlkZXIuXG4gICAgICpcbiAgICAgKiBAcGFyYW0gcHJvdmlkZXIgLSBUaGUgcHJvdmlkZXIgdG8gYWRkLlxuICAgICAqIEByZXR1cm5zIEEgbmV3bHkgYWxsb2NhdGVkIElEIGZvciB0aGUgcHJvdmlkZXIsIHVuaXF1ZSBhbW9uZyBhbGwgb3RoZXIgSURzIGluIHRoaXMgbWFwLCBhbmQgYW5cbiAgICAgKiAgICAgICAgICB1bnN1YnNjcmliYWJsZSBmb3IgdGhlIHByb3ZpZGVyLlxuICAgICAqIEB0aHJvd3MgSWYgdGhlcmUgYWxyZWFkeSBleGlzdHMgYW4gZW50cnkgd2l0aCB0aGUgZ2l2ZW4ge0BsaW5rIGlkfS5cbiAgICAgKi9cbiAgICBwdWJsaWMgYWRkKHByb3ZpZGVyOiBCKTogeyBpZDogbnVtYmVyOyBzdWJzY3JpcHRpb246IFVuc3Vic2NyaWJhYmxlIH0ge1xuICAgICAgICBjb25zdCBpZCA9IHRoaXMuaWRTZXF1ZW5jZVxuICAgICAgICB0aGlzLm1hcC5zZXQoaWQsIHByb3ZpZGVyKVxuICAgICAgICB0aGlzLmlkU2VxdWVuY2UrK1xuICAgICAgICByZXR1cm4geyBpZCwgc3Vic2NyaXB0aW9uOiB7IHVuc3Vic2NyaWJlOiAoKSA9PiB0aGlzLnJlbW92ZShpZCkgfSB9XG4gICAgfVxuXG4gICAgLyoqXG4gICAgICogUmV0dXJucyB0aGUgcHJvdmlkZXIgd2l0aCB0aGUgZ2l2ZW4ge0BsaW5rIGlkfS5cbiAgICAgKlxuICAgICAqIEB0ZW1wbGF0ZSBQIC0gVGhlIHNwZWNpZmljIHByb3ZpZGVyIHR5cGUgZm9yIHRoZSBwcm92aWRlciB3aXRoIHRoaXMge0BsaW5rIGlkfS5cbiAgICAgKiBAdGhyb3dzIElmIHRoZXJlIGlzIG5vIGVudHJ5IHdpdGggdGhlIGdpdmVuIHtAbGluayBpZH0uXG4gICAgICovXG4gICAgcHVibGljIGdldDxQIGV4dGVuZHMgQj4oaWQ6IG51bWJlcik6IFAge1xuICAgICAgICBjb25zdCBwcm92aWRlciA9IHRoaXMubWFwLmdldChpZCkgYXMgUFxuICAgICAgICBpZiAocHJvdmlkZXIgPT09IHVuZGVmaW5lZCkge1xuICAgICAgICAgICAgdGhyb3cgbmV3IEVycm9yKGBubyBwcm92aWRlciB3aXRoIElEICR7aWR9YClcbiAgICAgICAgfVxuICAgICAgICByZXR1cm4gcHJvdmlkZXJcbiAgICB9XG5cbiAgICAvKipcbiAgICAgKiBVbnN1YnNjcmliZXMgdGhlIHN1YnNjcmlwdGlvbiB0aGF0IHdhcyBwcmV2aW91c2x5IGFzc2lnbmVkIHRoZSBnaXZlbiB7QGxpbmsgaWR9LCBhbmQgcmVtb3ZlcyBpdCBmcm9tIHRoZVxuICAgICAqIG1hcC5cbiAgICAgKi9cbiAgICBwdWJsaWMgcmVtb3ZlKGlkOiBudW1iZXIpOiB2b2lkIHtcbiAgICAgICAgaWYgKCF0aGlzLm1hcC5oYXMoaWQpKSB7XG4gICAgICAgICAgICB0aHJvdyBuZXcgRXJyb3IoYG5vIHByb3ZpZGVyIHdpdGggSUQgJHtpZH1gKVxuICAgICAgICB9XG4gICAgICAgIHRyeSB7XG4gICAgICAgICAgICB0aGlzLnVuc3Vic2NyaWJlUHJvdmlkZXIoaWQpXG4gICAgICAgIH0gZmluYWxseSB7XG4gICAgICAgICAgICB0aGlzLm1hcC5kZWxldGUoaWQpXG4gICAgICAgIH1cbiAgICB9XG5cbiAgICAvKipcbiAgICAgKiBVbnN1YnNjcmliZXMgYWxsIHN1YnNjcmlwdGlvbnMgaW4gdGhpcyBtYXAgYW5kIGNsZWFycyBpdC5cbiAgICAgKi9cbiAgICBwdWJsaWMgdW5zdWJzY3JpYmUoKTogdm9pZCB7XG4gICAgICAgIHRyeSB7XG4gICAgICAgICAgICBmb3IgKGNvbnN0IGlkIG9mIHRoaXMubWFwLmtleXMoKSkge1xuICAgICAgICAgICAgICAgIHRoaXMudW5zdWJzY3JpYmVQcm92aWRlcihpZClcbiAgICAgICAgICAgIH1cbiAgICAgICAgfSBmaW5hbGx5IHtcbiAgICAgICAgICAgIHRoaXMubWFwLmNsZWFyKClcbiAgICAgICAgfVxuICAgIH1cbn1cbiJdfQ==