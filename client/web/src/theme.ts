import { useMemo } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { observeSystemIsLightTheme, ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/wildcard'

/**
 * The user preference for the theme.
 * These values are stored in temporary settings.
 */
export enum ThemePreference {
    Light = 'light',
    Dark = 'dark',
    System = 'system',
}
export interface ThemeState {
    /**
     * Parsed from temporary settings theme preference value.
     */
    themePreference: ThemePreference

    /**
     * Calculated theme preference. It value takes system preference
     * value into account if parsed value is equal to 'system'
     */
    enhancedThemePreference: ThemePreference.Light | ThemePreference.Dark

    setThemePreference: (theme: ThemePreference) => void
}

export const useTheme = (): ThemeState => {
    // React to system-wide theme change.
    const { observable: systemIsLightThemeObservable, initialValue: systemIsLightThemeInitialValue } = useMemo(
        () => observeSystemIsLightTheme(window),
        []
    )
    const systemIsLightTheme = useObservable(systemIsLightThemeObservable) ?? systemIsLightThemeInitialValue
    const [storedThemePreference, setThemePreference] = useTemporarySetting(
        'user.themePreference',
        ThemePreference.System
    )

    const themePreference = readStoredThemePreference(storedThemePreference)

    const enhancedThemePreference =
        themePreference === ThemePreference.System
            ? systemIsLightTheme
                ? ThemePreference.Light
                : ThemePreference.Dark
            : themePreference

    useMemo(() => {
        const isLightTheme = enhancedThemePreference === ThemePreference.Light

        document.documentElement.classList.toggle('theme-light', isLightTheme)
        document.documentElement.classList.toggle('theme-dark', !isLightTheme)
    }, [enhancedThemePreference])

    return {
        themePreference,
        enhancedThemePreference,
        setThemePreference,
    }
}

/**
 * Props that can be extended by any component's Props which needs to manipulate the theme preferences.
 *
 * @deprecated Use useTheme hook instead to get theme preference state
 */
export interface ThemePreferenceProps {
    themePreference: ThemePreference
    onThemePreferenceChange: (theme: ThemePreference) => void
}

/**
 * A React hook for getting and setting the theme.
 *
 * @deprecated Use useTheme hook instead to get theme preference state
 */
export const useThemeProps = (): ThemeProps & ThemePreferenceProps => {
    const { themePreference, enhancedThemePreference, setThemePreference } = useTheme()
    const isLightTheme = enhancedThemePreference === ThemePreference.Light

    return {
        isLightTheme,
        themePreference,
        onThemePreferenceChange: setThemePreference,
    }
}

/** Reads the stored theme preference from temporary settings */
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
