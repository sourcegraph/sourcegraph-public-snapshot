import { PublishedStoreItem } from '@storybook/client-api'
import { raw } from '@storybook/react'
import isChromatic from 'chromatic/isChromatic'

import { addStory } from './add-story'
import { storyStore } from './story-store'

// Execute logic below only in the environment where Chromatic snapshots are captured.
if (isChromatic()) {
    // CSF stories need to be evaluated before they are added to the `StoryStore` and thus are not immediately available.
    // We setTimeout to delay this logic until all stories have been added.
    setTimeout(() => {
        // Get an array of all stories which are already added to the `StoryStore`.
        // Use `raw()` because we don't want to apply any filtering and sorting on the array of stories.
        const storeItems = raw() as PublishedStoreItem[]

        // `StoryStore` is immutable outside of a configure() call.
        // As we delay this logic to support CSF stories, we need to set this to ensure changes are still applied.
        storyStore.startConfiguring()

        // Add three more versions of each story to test visual regressions with Chromatic snapshots.
        // In other environments, these themes can be explored by a user via toolbar toggles.
        for (const storeItem of storeItems) {
            // Default theme + Dark mode.
            addStory({ storeItem })
        }

        storyStore.finishConfiguring()
    }, 0)
}
