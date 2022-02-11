import isChromatic from 'chromatic/isChromatic'
const incrementedLocalStorageKeys = new Map<string, number>()

/**
 * Initializes a key in local storage with 0, then increments it at each new import of this module.
 * If the function is called in the same instance of this module over and over again, it'll keep returning
 * the same value (e.g. 0 after first initialization)
 * This is useful for incrementing a counter at each hard page load, but not at soft (React-level) reloads.
 *
 * "Shift" shifts the index of the page view to allow for alternating triggers.
 * E.g. if you have two uses of this hook, "A" with cadence = 4 and shift = 0, and "B" with cadence = 4 and shift = 2, you'll get this:
 *
 * > Page load #    Triggered
 * >     0              A
 * >     1              -
 * >     2              B
 * >     3              -
 * >     4              A
 * >     5              -
 * >    ...            ...
 *
 * It always returns `false` when running on Chromatic.
 */
export function usePersistentCadence(localStorageKey: string, cadence: number, shift: number = 0): boolean {
    if (isChromatic()) {
        return false
    }
    if (!incrementedLocalStorageKeys.has(localStorageKey)) {
        const pageViewCount = parseInt(localStorage.getItem(localStorageKey) || '', 10) || 0
        localStorage.setItem(localStorageKey, (pageViewCount + 1).toString())
        incrementedLocalStorageKeys.set(localStorageKey, pageViewCount)
        return pageViewCount % cadence === shift
    }

    const pageViewCount = incrementedLocalStorageKeys.get(localStorageKey) || 0
    return pageViewCount % cadence === shift
}

export function reset(): void {
    incrementedLocalStorageKeys.clear()
}
