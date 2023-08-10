import { useEffect, type RefObject } from 'react'

interface UseAutoFocusParameters {
    autoFocus?: boolean
    reference: RefObject<HTMLElement>
}

/**
 * Hook to ensure that an element is focused correctly.
 * Relying on the `autoFocus` attribute is not reliable within React.
 * https://reactjs.org/docs/accessibility.html#programmatically-managing-focus
 */
export const useAutoFocus = ({ autoFocus, reference }: UseAutoFocusParameters): void => {
    useEffect(() => {
        if (autoFocus) {
            requestAnimationFrame(() => {
                reference.current?.focus()
            })
        }
        // Reference will not change
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [autoFocus])
}
