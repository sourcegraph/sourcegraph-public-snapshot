const incrementedLocalStorageKeys = new Map<string, number>()

// Initializes a key in local storage with 0, then increments it at each new import of this module.
// If the function is called in the same instance of this module over and over again, it'll keep returning
// the same value (e.g. 0 after first initialization)
// This is useful for incrementing a counter at each hard page load, but not at soft (React-level) reloads.
export function usePersistentCadence(localStorageKey: string, cadence: number): boolean {
    if (!incrementedLocalStorageKeys.has(localStorageKey)) {
        const pageViewCount = parseInt(localStorage.getItem(localStorageKey) || '', 10) || 0
        localStorage.setItem(localStorageKey, (pageViewCount + 1).toString())
        incrementedLocalStorageKeys.set(localStorageKey, pageViewCount)
        return pageViewCount % cadence === 0
    }

    const pageViewCount = incrementedLocalStorageKeys.get(localStorageKey) || 0
    return pageViewCount % cadence === 0
}

export function reset(): void {
    incrementedLocalStorageKeys.clear()
}
