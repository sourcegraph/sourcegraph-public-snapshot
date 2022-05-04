import React, { ReactElement, useEffect } from 'react'

import { StoryFn } from '@storybook/addons'
import { useDarkMode } from 'storybook-dark-mode'

import { THEME_DARK_CLASS, THEME_LIGHT_CLASS } from '../themes'

export interface CreateChromaticStoryOptions {
    storyFn: StoryFn<ReactElement>
    isDarkModeEnabled: boolean
}

// Wrap `storyFn` into a decorator which takes care of CSS classes toggling based on received theme options.
export const createChromaticStory = (options: CreateChromaticStoryOptions): StoryFn => () => {
    const { storyFn, isDarkModeEnabled } = options
    // The `storyFn` is retrieved from the `StoryStore`, so it already has a `StoryContext`.
    // We can safely change its type to remove required props `StoryContext` props check.
    const Story = storyFn as React.ComponentType<React.PropsWithChildren<unknown>>

    const isDarkModeEnabledInitially = useDarkMode()

    useEffect(() => {
        // 'storybook-dark-mode' doesn't expose any API to toggle dark/light theme programmatically, so we do it manually.
        document.body.classList.toggle(THEME_DARK_CLASS, isDarkModeEnabled)
        document.body.classList.toggle(THEME_LIGHT_CLASS, !isDarkModeEnabled)
        document.body.dispatchEvent(new CustomEvent('chromatic-light-theme-toggled', { detail: !isDarkModeEnabled }))

        return () => {
            // Always toggle dark mode back to the previous value because otherwise, it might be out of sync with the toolbar toggle.
            document.body.classList.toggle(THEME_DARK_CLASS, isDarkModeEnabledInitially)
            document.body.classList.toggle(THEME_LIGHT_CLASS, !isDarkModeEnabledInitially)
            document.body.dispatchEvent(
                new CustomEvent('chromatic-light-theme-toggled', { detail: !isDarkModeEnabledInitially })
            )
        }
        // We need to execute `useEffect` callback once to take snapshot in Chromatic, so we can omit dependencies here.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    return <Story />
}
