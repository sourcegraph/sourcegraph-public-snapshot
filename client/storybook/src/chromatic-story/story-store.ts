import { StoryStore } from '@storybook/client-api'

// This global reference is used internally by Storybook:
// https://github.com/storybookjs/storybook/blob/3ec358f71c6111838092397d13fbe35b627a9a9d/lib/core-client/src/preview/start.ts#L43
declare global {
    interface Window {
        __STORYBOOK_STORY_STORE__: StoryStore
    }
}

// See the discussion about `StoryStore` usage in stories:
// https://github.com/storybookjs/storybook/discussions/12050#discussioncomment-125658
export const storyStore = window.__STORYBOOK_STORY_STORE__
