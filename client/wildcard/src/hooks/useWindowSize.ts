import { useState, useEffect } from 'react'

/**
 * Returns width and height of current window width.
 * Updates accordingly when window resizes.
 */
export function useWindowSize(): { width: number; height: number } {
    const [size, setSize] = useState({ height: window.innerHeight, width: window.innerWidth })

    useEffect(() => {
        function onResize(): void {
            setSize({ height: window.innerHeight, width: window.innerWidth })
        }
        window.addEventListener('resize', onResize)
        return () => window.removeEventListener('resize', onResize)
    }, [])

    return size
}
