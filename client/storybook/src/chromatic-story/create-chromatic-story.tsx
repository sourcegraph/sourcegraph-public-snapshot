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

export const createChromaticStory = (options: CreateChromaticStoryOptions): StoryFn => () => {
    const { storyFn, isRedesignEnabled, isDarkModeEnabled } = options
    // The storyFn is retrieved from the StoryStore, so it already has a StoryContext.
    const Story = storyFn as React.ComponentType

    const [_isRedesignEnabledInitially, setRedesignToggle] = useRedesignToggle()
    const isDarkModeEnabledInitially = useDarkMode()

    useEffect(() => {
        setRedesignToggle(isRedesignEnabled)
        // 'storybook-dark-mode' doesn't expose any method to toggle dark/light theme properly, so we do it manually.
        document.body.classList.toggle(THEME_DARK_CLASS, isDarkModeEnabled)
        document.body.classList.toggle(THEME_LIGHT_CLASS, !isDarkModeEnabled)

        return () => {
            // Do not enable redesign theme if it was disabled before this story was opened.
            if (isRedesignEnabled) {
                setRedesignToggle(!isRedesignEnabled)
            }
            document.body.classList.toggle(THEME_DARK_CLASS, isDarkModeEnabledInitially)
            document.body.classList.toggle(THEME_LIGHT_CLASS, !isDarkModeEnabledInitially)
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    return <Story />
}
