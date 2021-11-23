import { SetState } from 'zustand/vanilla'

/**
 * The user preference for the theme.
 * These values are stored in local storage.
 */
export enum ThemePreference {
    Light = 'light',
    Dark = 'dark',
    System = 'system',
}

export const LIGHT_THEME_LOCAL_STORAGE_KEY = 'light-theme'

/** Reads the stored theme preference from localStorage */
export const readStoredThemePreference = (value?: string): ThemePreference => {
    // Handle both old and new preference values
    switch (value) {
        case 'true':
        case 'light':
            return ThemePreference.Light
        case 'false':
        case 'dark':
            return ThemePreference.Dark
        default:
            return ThemePreference.System
    }
}

export interface ThemeState {
    theme: ThemePreference
    setTheme: (theme: ThemePreference) => void
}

export const createThemeStateStore = (set: SetState<ThemeState>): ThemeState => ({
    theme: readStoredThemePreference(),
    setTheme: theme => set({ theme }),
})
