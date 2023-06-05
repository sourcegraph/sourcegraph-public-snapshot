import { useCallback, useEffect, useMemo, useState } from 'react'

import { useLocation, useNavigate, useNavigationType } from 'react-router-dom'

export interface HistoryStack {
    size: number
    index: number
    canGoBack: boolean
    goBack: () => void
    canGoForward: boolean
    goForward: () => void
}

/**
 * Keep track of history stack size and index to determine if we can go back/forward
 * This is the only way of doing this on React Router v6 :(
 * See https://github.com/remix-run/react-router/discussions/8782#discussioncomment-2580895
 */
export function useHistoryStack(): HistoryStack {
    const [size, setSize] = useState(1)
    const [index, setIndex] = useState(1)

    const type = useNavigationType()
    const { key } = useLocation()

    // Update history stack size and index when navigation happens
    useEffect(() => {
        if (type === 'POP') {
            // Go back/forward in the stack, index is updated but the stack size stays the same
            // Don't update anything here, the index is already updated in goBack/goForward
        } else if (type === 'PUSH') {
            // Navigate to a new page, add it to the stack and set a new stack size
            const newIndex = index + 1
            setIndex(newIndex)
            setSize(newIndex)
        } else if (type === 'REPLACE') {
            // Replace the current page in the stack, index is the same but the stack size is reset
            setSize(index)
        }

        // Only run hook when type and path changes
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [type, key])

    const navigate = useNavigate()

    const canGoBack = useMemo(() => index > 1, [index])
    const goBack = useCallback(() => {
        if (canGoBack) {
            navigate(-1)
            setIndex(currentIndex => currentIndex - 1)
        }
    }, [canGoBack, navigate])

    const canGoForward = useMemo(() => index < size, [index, size])
    const goForward = useCallback(() => {
        if (canGoForward) {
            navigate(1)
            setIndex(currentIndex => currentIndex + 1)
        }
    }, [canGoForward, navigate])

    return { size, index, canGoBack, goBack, canGoForward, goForward }
}
