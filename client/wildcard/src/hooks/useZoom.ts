import { useState, useEffect } from 'react'

/**
 * Returns width and height of current window width.
 * Updates accordingly when window resizes.
 */
export const useZoom = (): number => {
    const [zoom, setZoom] = useState(window.devicePixelRatio)

    useEffect(() => {
        function onZoom(): void {
            setZoom(window.devicePixelRatio)
        }
        window.addEventListener('resize', onZoom)
        return () => window.removeEventListener('resize', onZoom)
    }, [])

    return zoom
}
