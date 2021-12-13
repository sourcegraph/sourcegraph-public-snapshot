import { useEffect, useState } from 'react'

/**
 * Returns provided media query match and, by default, subscribes to its updates.
 *
 * Example: `const isSmallScreen = useMatchMedia('(max-width: 250px)')`
 *
 * @param query CSS media query to match
 * @param observe Boolean flag indicating if you want to subscribe to updates, `true` by default
 * @returns `boolean`
 */
export function useMatchMedia(query: string, observe = true): boolean {
    const [isMatch, setIsMatch] = useState(window.matchMedia(query).matches)

    useEffect(() => {
        const handler = (event: MediaQueryListEvent): void => setIsMatch(event.matches)
        if (observe) {
            window.matchMedia(query).addEventListener('change', handler)
        }
        return () => window.matchMedia(query).removeEventListener('change', handler)
    }, [query, observe])

    return isMatch
}
