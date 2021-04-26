import { PublishedStoreItem } from '@storybook/client-api'
import { raw } from '@storybook/react'
import isChromatic from 'chromatic/isChromatic'

import { addStory } from './add-story'

if (isChromatic() || true) {
    // Get an array of all stories which are already added to the StoryStore.
    const stories = raw() as PublishedStoreItem[]

    // Add three more versions of each story to test visual regressions.
    // In other environments, these themes can be explored by a user via toolbar toggles.
    stories.map(storeItem => {
        // Default theme + Dark mode.
        addStory({
            storeItem,
            isDarkModeEnabled: true,
            isRedesignEnabled: false,
        })

        // Redesign theme + Light mode.
        addStory({
            storeItem,
            isDarkModeEnabled: false,
            isRedesignEnabled: true,
        })

        // Redesign theme + Dark mode.
        addStory({
            storeItem,
            isDarkModeEnabled: true,
            isRedesignEnabled: true,
        })
    })
}
