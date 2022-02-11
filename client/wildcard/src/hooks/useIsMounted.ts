import { useCallback, useEffect, useRef } from 'react'

/**
 * Custom hook which returns a function for checking if the component is still mounted.
 */
export const useIsMounted = (): (() => boolean) => {
    const isMounted = useRef(false)

    useEffect(() => {
        isMounted.current = true

        return () => {
            isMounted.current = false
        }
    }, [])

    return useCallback(() => isMounted.current, [])
}
