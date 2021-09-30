import { useEffect, useMemo } from 'react'
import create from 'zustand'

import { observeSystemIsLightTheme, ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

/**
 * The user preference for the theme.
 * These values are stored in local storage.
 */
export enum ThemePreference {
    Light = 'light',
    Dark = 'dark',
    System = 'system',
}

/**
 * Props that can be extended by any component's Props which needs to manipulate the theme preferences.
 */
export interface ThemePreferenceProps {
    themePreference: ThemePreference
    onThemePreferenceChange: (theme: ThemePreference) => void
}

const LIGHT_THEME_LOCAL_STORAGE_KEY = 'light-theme'

/** Reads the stored theme preference from localStorage */
const readStoredThemePreference = (localStorage: Pick<Storage, 'getItem' | 'setItem'>): ThemePreference => {
    const value = localStorage.getItem(LIGHT_THEME_LOCAL_STORAGE_KEY)
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

interface ThemeState {
    themePreference: ThemePreference
    onThemePreferenceChange: (themePreference?: ThemePreference) => void
}

export const useThemeStore = create<ThemeState>(set => ({
    themePreference: readStoredThemePreference(localStorage),
    onThemePreferenceChange: (preference?: ThemePreference) => {
        set(({ themePreference }) => {
            const newPreference =
                preference ?? (themePreference === ThemePreference.Dark ? ThemePreference.Light : ThemePreference.Dark)
            localStorage.setItem(LIGHT_THEME_LOCAL_STORAGE_KEY, newPreference)
            return { themePreference: newPreference }
        })
    },
}))

/**
 * A React hook for getting and setting the theme.
 *
 * @param window_ The global window object (or a mock in tests).
 * @param documentElement The root HTML document node (or a mock in tests).
 * @param localStorage The global localStorage object (or a mock in tests).
 */
export const useTheme = (
    window_: Pick<Window, 'matchMedia'> = window,
    documentElement: Pick<HTMLElement, 'classList'> = document.documentElement,
    localStorage: Pick<Storage, 'getItem' | 'setItem'> = window.localStorage
): ThemeProps & ThemePreferenceProps => {
    const { onThemePreferenceChange, themePreference } = useThemeStore()
    // React to system-wide theme change.
    const { observable: systemIsLightThemeObservable, initialValue: systemIsLightThemeInitialValue } = useMemo(
        () => observeSystemIsLightTheme(window_),
        [window_]
    )
    const systemIsLightTheme = useObservable(systemIsLightThemeObservable) ?? systemIsLightThemeInitialValue

    const isLightTheme = themePreference === 'system' ? systemIsLightTheme : themePreference === 'light'
    useEffect(() => {
        documentElement.classList.toggle('theme-light', isLightTheme)
        documentElement.classList.toggle('theme-dark', !isLightTheme)
    }, [documentElement.classList, isLightTheme])

    return {
        isLightTheme,
        themePreference,
        onThemePreferenceChange,
    }
}
