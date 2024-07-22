import type { Meta, StoryFn } from '@storybook/react'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'

import { WebStory } from '../../../components/WebStory'

import { SearchPageContent } from './SearchPageContent'

window.context.allowSignup = true

const config: Meta = {
    title: 'web/search/home/SearchPageContent',
    parameters: {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
        },
    },
}

export default config

export const CloudAuthedHome: StoryFn = () => (
    <WebStory
        legacyLayoutContext={{ isSourcegraphDotCom: true, authenticatedUser: { id: 'userID' } as AuthenticatedUser }}
    >
        {() => <SearchPageContent shouldShowAddCodeHostWidget={false} isSourcegraphDotCom={true} />}
    </WebStory>
)

CloudAuthedHome.storyName = 'Cloud authenticated home'

export const ServerHome: StoryFn = () => (
    <WebStory>{() => <SearchPageContent shouldShowAddCodeHostWidget={false} isSourcegraphDotCom={false} />}</WebStory>
)

ServerHome.storyName = 'Server home'
