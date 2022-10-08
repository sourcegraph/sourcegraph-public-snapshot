export type SourceboxTheme = SourceboxThemeColors | SourceboxBuiltinTheme

export interface SourceboxThemeColors {
    accent: string
    base: string
    surface1: string
}

export const SourceboxLightTheme: SourceboxThemeColors = {
    accent: '#888888',
    base: '#efefef',
    surface1: '#333333',
}

export const SourceboxDarkTheme: SourceboxThemeColors = {
    accent: '#bbbbbb',
    base: '#000000',
    surface1: '#eeeeee',
}

export type SourceboxBuiltinTheme = 'light' | 'dark'

export function resolveThemeColors(theme: SourceboxTheme): SourceboxThemeColors {
    if (theme === 'light') {
        return SourceboxLightTheme
    }
    if (theme === 'dark') {
        return SourceboxDarkTheme
    }
    return theme
}
