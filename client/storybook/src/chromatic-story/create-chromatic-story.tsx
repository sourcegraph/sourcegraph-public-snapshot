import { StoryFn } from '@storybook/addons'
import React, { ReactElement, useEffect } from 'react'
import { useDarkMode } from 'storybook-dark-mode'

import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

import { THEME_DARK_CLASS, THEME_LIGHT_CLASS } from '../themes'

export interface CreateChromaticStoryOptions {
    storyFn: StoryFn<ReactElement>
    isRedesignEnabled: boolean
    isDarkModeEnabled: boolean
}

// Wrap `storyFn` into a decorator which takes care of CSS classes toggling based on received theme options.
export const createChromaticStory = (options: CreateChromaticStoryOptions): StoryFn => () => {
    const { storyFn, isRedesignEnabled, isDarkModeEnabled } = options
    // The `storyFn` is retrieved from the `StoryStore`, so it already has a `StoryContext`.
    // We can safely change its type to remove required props `StoryContext` props check.
    const Story = storyFn as React.ComponentType

    const [, setRedesignToggle] = useRedesignToggle()
    const isDarkModeEnabledInitially = useDarkMode()

    useEffect(() => {
        setRedesignToggle(isRedesignEnabled)

        // 'storybook-dark-mode' doesn't expose any API to toggle dark/light theme programmatically, so we do it manually.
        document.body.classList.toggle(THEME_DARK_CLASS, isDarkModeEnabled)
        document.body.classList.toggle(THEME_LIGHT_CLASS, !isDarkModeEnabled)

        return () => {
            // Do not enable redesign theme if it was disabled before this story was opened.
            if (isRedesignEnabled) {
                setRedesignToggle(!isRedesignEnabled)
            }

            // Always toggle dark mode back to the previous value because otherwise, it might be out of sync with the toolbar toggle.
            document.body.classList.toggle(THEME_DARK_CLASS, isDarkModeEnabledInitially)
            document.body.classList.toggle(THEME_LIGHT_CLASS, !isDarkModeEnabledInitially)
        }
        // We need to execute `useEffect` callback once to take snapshot in Chromatic, so we can omit dependencies here.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    return <Story />
}
