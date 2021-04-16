import { useCallback, useEffect, useRef, useState } from 'react'

/**
 * A React hook to use and set state that is persisted in localStorage.
 * The getter's value will be updated whenever any setter globally is called.
 *
 * @param key The localStorage key to use for persistence.
 * @param initialValue The initial value to use when there is no value in localStorage for the key.
 * @param keepUpdated Whether to listen to update events from other useLocalStorage setters and keep the value up to date. (Default: false)
 * @returns A getter and setter for the value (`const [foo, setFoo] = useLocalStorage('key', 123)`).
 */
export const useLocalStorage = <T>(
    key: string,
    initialValue: T,
    { keepUpdated }: { keepUpdated: boolean } = { keepUpdated: false }
): [T, (value: T | ((previousValue: T) => T)) => void] => {
    const getCurrentValue = useCallback((): T => {
        try {
            const item = localStorage.getItem(key)
            return item ? (JSON.parse(item) as T) : initialValue
        } catch {
            return initialValue
        }
    }, [initialValue, key])

    const [storedValue, setStoredValue] = useState<T>(getCurrentValue)

    // We want `setValue` to have a stable identity like `setState`, so it shouldn't depend on `storedValue`.
    // Instead, have it read from a ref which is updated on each render.
    const storedValueReference = useRef<T>(storedValue)
    storedValueReference.current = storedValue

    const setValue = useCallback(
        (value: T | ((previousValue: T) => T)): void => {
            // We need to cast here because T could be a function type itself,
            // but we cannot tell TypeScript that functions are not allowed as T.
            const valueToStore =
                typeof value === 'function' ? (value as (previousValue: T) => T)(storedValueReference.current) : value
            window.localStorage.setItem(key, JSON.stringify(valueToStore))

            // This doesn't normally get fired on the same tab, needs to be fired
            // manually here so checking the event bellow works
            window.dispatchEvent(new Event('storage'))

            setStoredValue(valueToStore)
        },
        [key]
    )

    // Update value when localStorage changes
    useEffect(() => {
        if (!keepUpdated) {
            return
        }

        const updateOnStorageEvent = (): void => {
            const currentValue = getCurrentValue()
            setStoredValue(currentValue)
        }
        window.addEventListener('storage', updateOnStorageEvent)

        return () => window.removeEventListener('storage', updateOnStorageEvent)
    }, [getCurrentValue, keepUpdated])

    return [storedValue, setValue]
}
