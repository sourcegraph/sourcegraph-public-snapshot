import type { Meta, Story } from '@storybook/react'

import { WebStory } from '../components/WebStory'

import { RequestAccessPage } from './RequestAccessPage'

const config: Meta = {
    title: 'web/auth/RequestAccessPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default config

export const Default: Story = () => <WebStory>{() => <RequestAccessPage />}</WebStory>

export const Done: Story = () => (
    <WebStory initialEntries={[{ pathname: '/done' }]}>{() => <RequestAccessPage />}</WebStory>
)
