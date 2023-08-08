import { writable, type Unsubscriber, type Writable, type Readable } from 'svelte/store'

function isWritable<T>(value: any): value is Writable<T> {
    if (!value) {
        return false
    }
    return typeof value.subscribe === 'function' && typeof value.set === 'function'
}

interface WritableForwardStore<T> extends Writable<T> {
    updateStore(store: Writable<T>): void
}

interface ReadableForwardStore<T> extends Readable<T> {
    updateStore(store: Readable<T>): void
}

/**
 * Returns a helper store that syncs with the currently set store.
 */
export function createForwardStore<T>(store: Writable<T>): WritableForwardStore<T>
export function createForwardStore<T>(store: Readable<T>): ReadableForwardStore<T>
export function createForwardStore<T>(store: Writable<T> | Readable<T>) {
    const { subscribe, set } = writable<T>()
    let unsubscribe: Unsubscriber = store.subscribe(set)

    function link(store: Readable<T>): Unsubscriber {
        unsubscribe()
        return (unsubscribe = store.subscribe(set))
    }

    if (isWritable<T>(store)) {
        let writableStore = store

        return {
            subscribe,
            set(value) {
                writableStore.set(value)
            },
            update(value) {
                writableStore.update(value)
            },
            updateStore(newStore) {
                if (newStore !== writableStore) {
                    writableStore = newStore
                    link(writableStore)
                }
            },
        } satisfies WritableForwardStore<T>
    }

    return {
        subscribe,
        updateStore(newStore) {
            if (newStore !== store) {
                store = newStore
                link(store)
            }
        },
    } satisfies ReadableForwardStore<T>
}
