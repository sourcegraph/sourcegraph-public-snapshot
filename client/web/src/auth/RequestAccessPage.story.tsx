import type { Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../components/WebStory'

import { RequestAccessPage } from './RequestAccessPage'

const config: Meta = {
    title: 'web/auth/RequestAccessPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default config

export const Default: StoryFn = () => <WebStory>{() => <RequestAccessPage />}</WebStory>

export const Done: StoryFn = () => (
    <WebStory initialEntries={[{ pathname: '/done' }]}>{() => <RequestAccessPage />}</WebStory>
)
