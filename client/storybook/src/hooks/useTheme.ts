import { useLayoutEffect, useState } from 'react'

import { useDarkMode } from 'storybook-dark-mode'

/**
 * Gets current theme and updates value when theme changes
 *
 * @returns isLightTheme: boolean that is true if the light theme is enabled
 */
export const useTheme = (): boolean => {
    const isDarkMode = useDarkMode()
    const [isLightTheme, setIsLightTheme] = useState(!isDarkMode)

    // This is required for reacting to theme changes in local Storybook
    // via the toolbar button added by storybook-dark-mode
    useLayoutEffect(() => {
        setIsLightTheme(!isDarkMode)
    }, [isDarkMode])

    // This is required for Chromatic to react to theme changes when
    // taking screenshots. See `create-chromatic-story.tsx` where
    // this event is dispatched.
    useLayoutEffect(() => {
        const listener = ((event: CustomEvent<boolean>): void => {
            setIsLightTheme(event.detail)
        }) as EventListener
        document.body.addEventListener('chromatic-light-theme-toggled', listener)
        return () => document.body.removeEventListener('chromatic-light-theme-toggled', listener)
    }, [])

    return isLightTheme
}
