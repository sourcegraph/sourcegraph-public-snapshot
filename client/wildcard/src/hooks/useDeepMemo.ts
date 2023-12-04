import { useRef } from 'react'

import { isEqual } from 'lodash'

/**
 * Returns memoized value that is checked with lodash deep memo
 * equal helper.
 */
export function useDeepMemo<Value>(value: Value): Value {
    const previousValueReference = useRef<Value>(value)

    if (!isEqual(previousValueReference.current, value)) {
        previousValueReference.current = value
    }

    return previousValueReference.current
}
