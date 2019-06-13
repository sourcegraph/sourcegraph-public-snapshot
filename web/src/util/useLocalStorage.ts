import { useState } from 'react'

/**
 * A React hook to use and set state that is persisted in localStorage.
 *
 * @param key The localStorage key to use for persistence.
 * @param initialValue The initial value to use when there is no value in localStorage for the key.
 * @returns A getter and setter for the value (`const [foo, setFoo] = useLocalStorage('key', 123)`).
 */
export const useLocalStorage = <T>(key: string, initialValue: T): [T, (value: T) => void] => {
    const [storedValue, setStoredValue] = useState<T>(() => {
        try {
            const item = localStorage.getItem(key)
            return item ? JSON.parse(item) : initialValue
        } catch (error) {
            return initialValue
        }
    })

    const setValue = (value: T): void => {
        const valueToStore = typeof value === 'function' ? value(storedValue) : value
        setStoredValue(valueToStore)
        window.localStorage.setItem(key, JSON.stringify(valueToStore))
    }

    return [storedValue, setValue]
}
