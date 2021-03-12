import { useState } from 'react'

interface UseContolledParameters<T> {
    onChange?: (value: T) => void
    value: T
}

export function useControlledState<T>({ value, onChange }: UseContolledParameters<T>): [T, (item: T) => void] {
    const [uncontrolledValue, setUncontrolledValue] = useState(value)

    // State is already controlled
    if (onChange) {
        return [value, onChange]
    }

    // We must control the state ourselves
    return [uncontrolledValue, setUncontrolledValue]
}
