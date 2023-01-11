import { useEffect, useRef } from 'react'

/**
 * This hook is the same as `useEffect` but ignore the first call for `didMount`
 */
export const useUpdateEffect: typeof useEffect = (effect, deps): void => {
    const isFirst = useRef(true)

    useEffect(() => {
        if (!isFirst.current) {
            return effect()
        }

        isFirst.current = false

        // We only care about `deps` for this hook usages
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, deps)
}
