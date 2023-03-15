import { createContext, useContext } from 'react'

export interface ChromaticTheme {
    theme: 'light' | 'dark'
}

export const ChromaticThemeContext = createContext<ChromaticTheme>({
    theme: 'light',
})

export function useChromaticDarkMode(): boolean {
    return useContext(ChromaticThemeContext).theme === 'dark'
}
