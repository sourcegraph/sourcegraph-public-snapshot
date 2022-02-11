import { createContext, useContext } from 'react'

export interface WildcardTheme {
    isBranded: boolean
}

export const WildcardThemeContext = createContext<WildcardTheme>({
    isBranded: false,
})

export function useWildcardTheme(): WildcardTheme {
    return useContext(WildcardThemeContext)
}
