import { Dispatch, SetStateAction, useCallback, useRef, useState } from 'react'

/**
 * A helper method to convert any `Storage` object into a React hook, such as `useLocalStorage`.
 */
export const useStorageHook = <T>(storage: Storage, key: string, initialValue: T): [T, Dispatch<SetStateAction<T>>] => {
    const [storedValue, setStoredValue] = useState<T>(() => {
        try {
            const item = storage.getItem(key)
            return item ? (JSON.parse(item) as T) : initialValue
        } catch {
            return initialValue
        }
    })

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
            setStoredValue(valueToStore)
        },
        [storage, key]
    )

    return [storedValue, setValue]
}
