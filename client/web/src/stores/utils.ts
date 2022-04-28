import { StateSelector, UseBoundStore } from 'zustand'

/**
 * Helper function for creating a hook that extract a single value from a store.
 */
export function createSingle<T extends object, U = unknown>(
    store: UseBoundStore<T>,
    selector: StateSelector<T, U>
): () => U {
    return () => store(selector)
}
