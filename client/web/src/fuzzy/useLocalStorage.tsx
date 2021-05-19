import { useState } from 'react'

import { useLocalStorage as baseUseLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

export interface State<T> {
    value: T
    set(newValue: T): void
}

/**
 * Wrapper around React `useState` that returns `State<T>
 */
export function useEphemeralState<T>(initialValue: T): State<T> {
    const [value, set] = useState(initialValue)
    return { value, set }
}

/**
 * Wrapper around React `useState` that returns `State<T> and caches the result in window.localStorage.
 */
export function useLocalStorage<T>(key: string, initialValue: T): State<T> {
    const [value, set] = baseUseLocalStorage(key, initialValue)
    return { value, set }
}
