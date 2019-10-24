/**
 * Props that can be extended by any component's Props which needs to react to theme change.
 */
export interface ThemeProps {
    /**
     * `true` if the current theme to be shown is the light theme,
     * `false` if it is the dark theme.
     */
    isLightTheme: boolean
}

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
