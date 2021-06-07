import { PublishedStoreItem } from '@storybook/client-api'
import { raw } from '@storybook/react'
import isChromatic from 'chromatic/isChromatic'

import { addStory } from './add-story'

// Execute logic below only in the environment where Chromatic snapshots are captured.
if (isChromatic()) {
    // Get an array of all stories which are already added to the `StoryStore`.
    // Use `raw()` because we don't want to apply any filtering and sorting on the array of stories.
    const storeItems = raw() as PublishedStoreItem[]

    // Add three more versions of each story to test visual regressions with Chromatic snapshots.
    // In other environments, these themes can be explored by a user via toolbar toggles.
    for (const storeItem of storeItems) {
        // Default theme + Dark mode.
        addStory({
            storeItem,
            isDarkModeEnabled: true,
        })
    }
}
