import { get, writable, type Unsubscriber, type Writable } from 'svelte/store'

/**
 * Returns a helper store that syncs with the currently set store.
 */
export function createForwardStore<T>(store: Writable<T>): Writable<T> & { updateStore(store: Writable<T>): void } {
    const { subscribe, set } = writable<T>(get(store), () => link(store))

    let unsubscribe: Unsubscriber | null = null
    function link(store: Writable<T>): Unsubscriber {
        unsubscribe?.()
        return (unsubscribe = store.subscribe(set))
    }

    return {
        subscribe,
        set(value) {
            store.set(value)
        },
        update(value) {
            store.update(value)
        },
        updateStore(newStore) {
            if (newStore !== store) {
                store = newStore
                link(store)
            }
        },
    }
}
