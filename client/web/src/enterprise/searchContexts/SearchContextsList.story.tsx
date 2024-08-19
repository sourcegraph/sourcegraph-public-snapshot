import type { Decorator, Meta, StoryFn } from '@storybook/react'

import {
    mockAuthenticatedUser,
    mockFetchSearchContexts,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { WebStory } from '../../components/WebStory'

import { SearchContextsList, type SearchContextsListProps } from './SearchContextsList'

const decorator: Decorator = story => (
    <div className="p-3 container" style={{ position: 'static' }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'web/enterprise/searchContexts/SearchContextsListTab',
    decorators: [decorator],
    parameters: {},
}

export default config

const defaultProps: SearchContextsListProps = {
    authenticatedUser: mockAuthenticatedUser,
    fetchSearchContexts: mockFetchSearchContexts,
    platformContext: NOOP_PLATFORM_CONTEXT,
    setAlert: () => undefined,
}

export const Default: StoryFn = () => <WebStory>{() => <SearchContextsList {...defaultProps} />}</WebStory>
