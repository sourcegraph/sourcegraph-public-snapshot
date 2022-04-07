import { useRef } from 'react'

import { isEqual } from 'lodash'

export function useDistinctValue<Value>(value: Value): Value {
    const previousValueReference = useRef<Value>(value)

    if (!isEqual(previousValueReference.current, value)) {
        previousValueReference.current = value
    }

    return previousValueReference.current
}
