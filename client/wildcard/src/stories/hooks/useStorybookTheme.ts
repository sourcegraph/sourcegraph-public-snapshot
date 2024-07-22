import { useLayoutEffect, useState } from 'react'

import { useDarkMode } from 'storybook-dark-mode'

/**
 * Gets current theme and updates value when theme changes
 * @returns isLightTheme: boolean that is true if the light theme is enabled
 */
export const useStorybookTheme = (): boolean => {
    const isDarkMode = useDarkMode()
    const [isLightTheme, setIsLightTheme] = useState(!isDarkMode)

    // This is required for reacting to theme changes in local Storybook
    // via the toolbar button added by storybook-dark-mode
    useLayoutEffect(() => {
        setIsLightTheme(!isDarkMode)
    }, [isDarkMode])

    return isLightTheme
}
