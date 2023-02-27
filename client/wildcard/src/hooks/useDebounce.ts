import { useState, useEffect } from 'react'

/**
 * This function will trail debounce a changing value
 *
 * @param value The value expected to change
 * @param delay Delay before updating the value
 * @returns The updated value
 */
export const useDebounce = <T>(value: T, delay: number): T => {
    const [debouncedValue, setDebouncedValue] = useState(value)

    useEffect(() => {
        const handler = setTimeout(() => setDebouncedValue(value), delay)
        return () => clearTimeout(handler)
    }, [delay, value])

    return delay === 0 ? value : debouncedValue
}
