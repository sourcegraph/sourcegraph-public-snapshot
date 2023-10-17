import type { ReactElement } from 'react'

import type { Decorator } from '@storybook/react'

import { ChromaticRoot } from './ChromaticRoot'

/**
 * The global Storybook decorator used to snapshot stories with multiple themes in Chromatic.
 *
 * It's a recommended way of achieving this goal:
 * https://www.chromatic.com/docs/faq#do-you-support-taking-snapshots-of-a-component-with-multiple-the
 *
 * If the `chromatic.enableDarkMode` story parameter is set to `true`, the story will
 * be rendered twice in Chromatic — in light and dark modes.
 */
export const withChromaticThemes: Decorator<ReactElement> = (StoryFunc, { parameters }) => {
    if (parameters?.chromatic?.enableDarkMode) {
        return (
            <>
                <ChromaticRoot theme="light">
                    <StoryFunc />
                </ChromaticRoot>

                <ChromaticRoot theme="dark">
                    <StoryFunc />
                </ChromaticRoot>
            </>
        )
    }

    return <StoryFunc />
}
