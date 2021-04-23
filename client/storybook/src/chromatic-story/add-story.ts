import { toId } from '@storybook/csf'
import { PublishedStoreItem, StoryStore } from '@storybook/client-api'

import { createChromaticStory, CreateChromaticStoryOptions } from './create-chromatic-story'

// See https://github.com/storybookjs/storybook/discussions/12050#discussioncomment-125658
declare global {
    interface Window {
        __STORYBOOK_STORY_STORE__: StoryStore
    }
}

const storyStore = window.__STORYBOOK_STORY_STORE__

interface AddStoryOptions extends Pick<CreateChromaticStoryOptions, 'isRedesignEnabled' | 'isDarkModeEnabled'> {
    storeItem: PublishedStoreItem
}

export const addStory = (options: AddStoryOptions) => {
    const {
        storeItem: { name, kind, storyFn, parameters },
        isDarkModeEnabled,
        isRedesignEnabled,
    } = options

    const storyName = [name, isRedesignEnabled && '[Redesign]', isDarkModeEnabled && '🌚'].filter(Boolean).join(' ')

    // Use storyStore.addStory to avoid applying decorators for stories, because PublishedStoreItem.storyFn already has decorators applied.
    storyStore.addStory(
        {
            id: toId(kind, storyName),
            kind,
            name: storyName,
            parameters,
            loaders: [],
            storyFn: createChromaticStory({
                storyFn,
                isDarkModeEnabled,
                isRedesignEnabled,
            }),
        },
        {
            applyDecorators: storyFn => storyFn,
        }
    )
}
