import { useState } from 'react'

interface UseContolledParameters<T> {
    onChange?: (value: T) => void
    value: T
}

/**
 * A hook to allow other components & hooks to easily support both controlled and uncontrolled variations of state.
 * `useControlledState` acts like `useState` except it assumes it can defer state management to the caller if an `onChange` parameter is passed.
 */
export function useControlledState<T>({ value, onChange }: UseContolledParameters<T>): [T, (item: T) => void] {
    const [uncontrolledValue, setUncontrolledValue] = useState(value)

    // State is already controlled
    if (onChange) {
        return [value, onChange]
    }

    // We must control the state ourselves
    return [uncontrolledValue, setUncontrolledValue]
}
