import type { Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { SiteInitPage } from './SiteInitPage'

const config: Meta = {
    title: 'web/auth/SiteInitPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default config

export const Default: StoryFn = () => (
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

export const Authenticated: StoryFn = () => (
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
