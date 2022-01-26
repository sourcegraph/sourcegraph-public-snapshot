import { useMemo } from 'react'

import { observeSystemIsLightTheme, ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/wildcard'

import { useThemeState } from './stores'
import { ThemePreference } from './stores/themeState'

export interface ThemeState {
    /**
     * Parsed from local storage theme preference value.
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

    const [themePreference, setThemePreference] = useThemeState(state => [state.theme, state.setTheme])
    const enhancedThemePreference =
        themePreference === ThemePreference.System
            ? systemIsLightTheme
                ? ThemePreference.Light
                : ThemePreference.Dark
            : themePreference

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

    useMemo(() => {
        document.documentElement.classList.toggle('theme-light', isLightTheme)
        document.documentElement.classList.toggle('theme-dark', !isLightTheme)
    }, [isLightTheme])

    return {
        isLightTheme,
        themePreference,
        onThemePreferenceChange: setThemePreference,
    }
}
