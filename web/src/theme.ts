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
    HighContrast = 'hc-black',
    System = 'system',
}

/**
 * The actual themes that are available. The "System" theme is not an actual theme itself, so it's
 * omitted.
 */
export type Theme = Exclude<ThemePreference, ThemePreference.System>

/**
 * Props that can be extended by any component's Props which needs to manipulate the theme preferences.
 */
export interface ThemePreferenceProps {
    themePreference: ThemePreference
    onThemePreferenceChange: (theme: ThemePreference) => void
}
