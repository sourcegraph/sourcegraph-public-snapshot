import { useState } from 'react'

export function useLocalStorage<T>(key: string, initialValue: T): [T, (value: T) => void] {
    const [storedValue, setStoredValue] = useState<T>(() => {
        try {
            const item = window.localStorage.getItem(key)
            return item ? (JSON.parse(item) as T) : initialValue
        } catch {
            return initialValue
        }
    })

    const setValue = (value: T | ((value_: T) => T)): void => {
        const valueToStore = value instanceof Function ? value(storedValue) : value
        setStoredValue(valueToStore)
        window.localStorage.setItem(key, JSON.stringify(valueToStore))
    }

    return [storedValue, setValue]
}
