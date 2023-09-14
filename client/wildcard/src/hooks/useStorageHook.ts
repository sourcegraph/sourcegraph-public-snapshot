import { type Dispatch, type SetStateAction, useCallback, useRef, useSyncExternalStore, useMemo } from 'react'

import { isEqual } from 'lodash'

/**
 * Since storage doesn't emit value to the current tab we should be able
 * to trigger use sync external hook outside subscribe function, for example
 * in setState function in order to update the current tab state.
 */
const callbacks = new Map<string, Set<() => void>>()

function addCallback(key: string, callback: () => void): void {
    if (!callbacks.has(key)) {
        callbacks.set(key, new Set([callback]))
        return
    }

    const set = callbacks.get(key)
    set?.add(callback)
}

function removeCallback(key: string, callback: () => void): void {
    if (!callbacks.has(key)) {
        return
    }

    const set = callbacks.get(key)
    set?.delete(callback)
}

function notifyCallbacks(key: string): void {
    const set = callbacks.get(key)

    if (set) {
        for (const callback of set.values()) {
            // eslint-disable-next-line callback-return
            callback()
        }
    }
}

/**
 * A helper method to convert any `Storage` object into a React hook, such as `useLocalStorage`.
 */
export const useStorageHook = <T>(storage: Storage, key: string, initialValue: T): [T, Dispatch<SetStateAction<T>>] => {
    const subscribe = useMemo(() => subscribeToStorage(key), [key])
    const getSnapshot = useMutableSnapshot<T>({ key, storage, initialValue })

    const storedValue = useSyncExternalStore<T>(subscribe, getSnapshot)

    const valueRef = useRef(storedValue)
    valueRef.current = storedValue

    const setValue: Dispatch<SetStateAction<T>> = useCallback(
        (value: T | ((previousValue: T) => T)): void => {
            // We need to cast here because T could be a function type itself,
            // but we cannot tell TypeScript that functions are not allowed as T.
            const valueToStore =
                typeof value === 'function' ? (value as (previousValue: T) => T)(valueRef.current) : value
            storage.setItem(key, JSON.stringify(valueToStore))
            notifyCallbacks(key)
        },
        [storage, key]
    )

    return [storedValue, setValue]
}

type Unsubscribe = () => void
type Subscribe = (onStoreChange: () => void) => Unsubscribe

function subscribeToStorage(key: string): Subscribe {
    return (callback: () => void): Unsubscribe => {
        function onStorageKeyChange(event: StorageEvent): void {
            if (event.key === key) {
                callback()
                return
            }
        }

        // Register sync external store callback to trigger it later
        // outside subscribe function
        addCallback(key, callback)

        // Listen storage change events (changes from other browser tabs)
        addEventListener('storage', onStorageKeyChange)

        return () => {
            removeCallback(key, callback)
            removeEventListener('storage', onStorageKeyChange)
        }
    }
}

interface UseMutableSnapshotProps<T> {
    key: string
    storage: Storage
    initialValue: T
}

function useMutableSnapshot<T>(props: UseMutableSnapshotProps<T>): () => T {
    const { key, storage, initialValue } = props

    const previousValueReference = useRef<T>(initialValue)

    return useCallback(() => {
        const newValue = getStorageValue<T>(storage, key) ?? initialValue

        if (!isEqual(newValue, previousValueReference.current)) {
            previousValueReference.current = newValue
        }

        return previousValueReference.current
    }, [storage, key, initialValue])
}

function getStorageValue<T>(storage: Storage, key: string): T | null {
    try {
        const item = storage.getItem(key)
        return item ? (JSON.parse(item) as T) : null
    } catch {
        return null
    }
}
