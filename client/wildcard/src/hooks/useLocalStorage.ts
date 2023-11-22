import type { Dispatch, SetStateAction } from 'react'

import { useStorageHook } from './useStorageHook'

/**
 * A React hook to use and set state that is persisted in localStorage.
 *
 * @param key The localStorage key to use for persistence.
 * @param initialValue The initial value to use when there is no value in localStorage for the key.
 * @returns A getter and setter for the value (`const [foo, setFoo] = useLocalStorage('key', 123)`).
 */
export const useLocalStorage = <T>(key: string, initialValue: T): [T, Dispatch<SetStateAction<T>>] =>
    useStorageHook(window.localStorage, key, initialValue)
