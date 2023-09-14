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

/**
 * Returns `true` if the user has opted for reduced motion.
 *
 * Using `no-preference` instead of `reduce` here is better because if the user
 * is using a browser that does not support reduced motion, the media query
 * will not match and the user will get the reduced motion experience,
 * which is the safer choice.
 */
export function useReducedMotion(): boolean {
    return !useMatchMedia('(prefers-reduced-motion: no-preference)')
}
