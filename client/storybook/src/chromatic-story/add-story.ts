import { PublishedStoreItem } from '@storybook/client-api'
import { toId } from '@storybook/csf'

import { createChromaticStory } from './create-chromatic-story'
import { storyStore } from './story-store'

interface AddStoryOptions {
    storeItem: PublishedStoreItem
}

export const addStory = (options: AddStoryOptions): void => {
    const {
        storeItem: { name, kind, storyFn, parameters },
    } = options

    const isDarkModeEnabled = Boolean(parameters?.chromatic?.enableDarkMode)

    // Add suffix to the story name based on theme options:
    // 1. Default + Dark:   "Text" -> "Text 🌚"
    const storyName = [name, isDarkModeEnabled && '🌚'].filter(Boolean).join(' ')

    /**
     * Use `storyStore.addStory()` to avoid applying decorators to stories, because `PublishedStoreItem.storyFn` already has decorators applied.
     * `storiesOf().add()` usage API would result in decorators duplication. It's possible to avoid this issue using `PublishedStoreItem.getOriginal()`,
     * which returns only story function without any decorators and story context. It means that we should apply them manually and
     * keep this logic in sync with Storybook internals to have consistent behavior. `storyStore.addStory()` allows to avoid it.
     */
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
            }),
        },
        {
            // The default `applyDecorators` implementation accepts `decorators` as a second arg and applies them to the `storyFn`.
            // Our `storyFn` already has all the decorators applied, so we just return it.
            applyDecorators: storyFunc => storyFunc,
        }
    )
}
