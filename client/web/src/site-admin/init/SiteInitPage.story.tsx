import type { Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { SiteInitPage } from './SiteInitPage'

const config: Meta = {
    title: 'web/auth/SiteInitPage',
}

export default config

export const Default: Story = () => (
    <WebStory>
        {() => (
            <SiteInitPage
                context={{
                    authMinPasswordLength: 12,
                    authPasswordPolicy: {},
                }}
                authenticatedUser={null}
                needsSiteInit={true}
            />
        )}
    </WebStory>
)

export const Authenticated: Story = () => (
    <WebStory>
        {() => (
            <SiteInitPage
                context={{
                    authMinPasswordLength: 12,
                    authPasswordPolicy: {},
                }}
                authenticatedUser={{ username: 'johndoe' }}
                needsSiteInit={true}
            />
        )}
    </WebStory>
)
