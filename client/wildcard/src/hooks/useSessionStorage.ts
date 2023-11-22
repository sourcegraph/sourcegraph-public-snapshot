import type { Dispatch, SetStateAction } from 'react'

import { useStorageHook } from './useStorageHook'

/**
 * A React hook to use and set state that is persisted in sessionStorage.
 *
 * @param key The sessionStorage key to use for persistence.
 * @param initialValue The initial value to use when there is no value in sessionStorage for the key.
 * @returns A getter and setter for the value (`const [foo, setFoo] = useSessionStorage('key', 123)`).
 */
export const useSessionStorage = <T>(key: string, initialValue: T): [T, Dispatch<SetStateAction<T>>] =>
    useStorageHook(window.sessionStorage, key, initialValue)
