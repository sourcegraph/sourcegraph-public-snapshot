import { DecoratorFn, Meta, Story } from '@storybook/react'

import {
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { WebStory } from '../../components/WebStory'

import { SearchContextsList, SearchContextsListProps } from './SearchContextsList'

const decorator: DecoratorFn = story => (
    <div className="p-3 container" style={{ position: 'static' }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'web/enterprise/searchContexts/SearchContextsListTab',
    decorators: [decorator],
    parameters: {
        chromatic: { viewports: [1200], disableSnapshot: false },
    },
}

export default config

const defaultProps: SearchContextsListProps = {
    authenticatedUser: null,
    fetchSearchContexts: mockFetchSearchContexts,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    platformContext: NOOP_PLATFORM_CONTEXT,
}

export const Default: Story = () => <WebStory>{() => <SearchContextsList {...defaultProps} />}</WebStory>
