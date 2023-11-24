import { derived, writable, type Readable } from 'svelte/store'

import { createMappingStore } from './utils'

export enum Theme {
    Light,
    Dark,
    System,
}

/**
 * The currently selected Theme.
 */
export const theme = writable(Theme.System)

/**
 * This store returns true if the theme is set to light or if the user's system
 * preference is 'light'. The store listens to match media changes and updates
 * accordingly.
 */
export const isLightTheme = derived(theme, ($theme, set) => {
    if ($theme === Theme.System) {
        const matchMedia = window.matchMedia('(prefers-color-scheme: light)')
        set(matchMedia.matches)
        const listener = (event: MediaQueryListEventMap['change']) => {
            set(event.matches)
        }
        matchMedia.addEventListener('change', listener)
        return () => matchMedia.removeEventListener('change', listener)
    }
    set($theme === Theme.Light)
    return
}) satisfies Readable<boolean>

/**
 * A store that maps user friendly theme names to Theme.* values and vice versa.
 */
export const humanTheme = createMappingStore(
    theme,
    theme => {
        switch (theme) {
            case Theme.Light: {
                return 'Light'
            }
            case Theme.Dark: {
                return 'Dark'
            }
            case Theme.System: {
                return 'System'
            }
        }
    },
    value => {
        switch (value) {
            case 'Light': {
                return Theme.Light
            }
            case 'Dark': {
                return Theme.Dark
            }
            case 'System': {
                return Theme.System
            }
        }
    }
)
