import { Dispatch, SetStateAction, useCallback, useRef, useSyncExternalStore, useMemo } from 'react'

/**
 * A helper method to convert any `Storage` object into a React hook, such as `useLocalStorage`.
 */
export const useStorageHook = <T>(storage: Storage, key: string, initialValue: T): [T, Dispatch<SetStateAction<T>>] => {
    const subscribe = useMemo(() => subscribeToStorage(key), [key])
    const getSnapshot = useCallback(() => getStorageValue(storage, key, initialValue), [storage, key, initialValue])

    const storedValue = useSyncExternalStore(subscribe, getSnapshot)

    // We want `setValue` to have a stable identity like `setState`, so it shouldn't depend on `storedValue`.
    // Instead, have it read from a ref which is updated on each render.
    const storedValueReference = useRef<T>(storedValue)
    storedValueReference.current = storedValue

    const setValue: Dispatch<SetStateAction<T>> = useCallback(
        (value: T | ((previousValue: T) => T)): void => {
            // We need to cast here because T could be a function type itself,
            // but we cannot tell TypeScript that functions are not allowed as T.
            const valueToStore =
                typeof value === 'function' ? (value as (previousValue: T) => T)(storedValueReference.current) : value
            storage.setItem(key, JSON.stringify(valueToStore))
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

        addEventListener('storage', onStorageKeyChange)

        return () => {
            removeEventListener('storage', onStorageKeyChange)
        }
    }
}

function getStorageValue<T>(storage: Storage, key: string, fallbackValue: T): T {
    try {
        const item = storage.getItem(key)
        return item ? (JSON.parse(item) as T) : fallbackValue
    } catch {
        return fallbackValue
    }
}
