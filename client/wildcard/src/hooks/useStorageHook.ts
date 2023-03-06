import { Dispatch, SetStateAction, useCallback, useRef, useSyncExternalStore, useMemo } from 'react'

/**
 * A helper method to convert any `Storage` object into a React hook, such as `useLocalStorage`.
 */
export const useStorageHook = <T>(storage: Storage, key: string, initialValue: T): [T, Dispatch<SetStateAction<T>>] => {
    const subscribe = useMemo(() => subscribeToStorage(key), [key])
    const getSnapshot = useMutableSnapshot<T>({ key, storage, initialValue })

    const storedValue = useSyncExternalStore<{ value: T }>(subscribe, getSnapshot)

    const setValue: Dispatch<SetStateAction<T>> = useCallback(
        (value: T | ((previousValue: T) => T)): void => {
            // We need to cast here because T could be a function type itself,
            // but we cannot tell TypeScript that functions are not allowed as T.
            const valueToStore =
                typeof value === 'function' ? (value as (previousValue: T) => T)(storedValue.value) : value
            storage.setItem(key, JSON.stringify(valueToStore))
        },
        [storage, key, storedValue]
    )

    return [storedValue.value, setValue]
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

interface UseMutableSnapshotProps<T> {
    key: string
    storage: Storage
    initialValue: T
}

function useMutableSnapshot<T>(props: UseMutableSnapshotProps<T>): () => { value: T } {
    const { key, storage, initialValue } = props
    const mutableValue = useRef<{ value: T }>({ value: initialValue })

    return useCallback(() => {
        const newValue = getStorageValue<T>(storage, key)
        mutableValue.current.value = newValue ?? initialValue

        return mutableValue.current
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
