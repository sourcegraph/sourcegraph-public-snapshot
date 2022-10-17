import { DecoratorFn, Meta, Story } from '@storybook/react'
import { subDays } from 'date-fns'
import { Observable, of } from 'rxjs'

import { ListSearchContextsResult } from '@sourcegraph/search'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { WebStory } from '../../components/WebStory'

import { SearchContextsListTab, SearchContextsListTabProps } from './SearchContextsListTab'

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

const defaultProps: SearchContextsListTabProps = {
    authenticatedUser: null,
    isSourcegraphDotCom: true,
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
    fetchSearchContexts: mockFetchSearchContexts,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    platformContext: NOOP_PLATFORM_CONTEXT,
}

const propsWithContexts: SearchContextsListTabProps = {
    ...defaultProps,
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(1),
    fetchSearchContexts: ({
        first,
        query,
        after,
    }: {
        first: number
        query?: string
        after?: string
    }): Observable<ListSearchContextsResult['searchContexts']> =>
        of({
            nodes: [
                {
                    __typename: 'SearchContext',
                    id: '3',
                    spec: '@username/test-version-1.5',
                    name: 'test-version-1.5',
                    namespace: {
                        __typename: 'User',
                        id: 'u1',
                        namespaceName: 'username',
                    },
                    autoDefined: false,
                    public: true,
                    description: 'Only code in version 1.5',
                    query: '',
                    updatedAt: subDays(new Date(), 1).toISOString(),
                    repositories: [],
                    viewerCanManage: true,
                },
                {
                    __typename: 'SearchContext',
                    id: '4',
                    spec: '@username/test-version-1.6',
                    namespace: {
                        __typename: 'User',
                        id: 'u1',
                        namespaceName: 'username',
                    },
                    name: 'test-version-1.6',
                    autoDefined: false,
                    public: false,
                    description: 'Only code in version 1.6',
                    query: '',
                    updatedAt: subDays(new Date(), 1).toISOString(),
                    repositories: [],
                    viewerCanManage: true,
                },
            ],
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            totalCount: 1,
        }),
}

export const Default: Story = () => <WebStory>{() => <SearchContextsListTab {...defaultProps} />}</WebStory>

export const WithSourcegraphDotComDisabled: Story = () => (
    <WebStory>{() => <SearchContextsListTab {...propsWithContexts} isSourcegraphDotCom={false} />}</WebStory>
)

WithSourcegraphDotComDisabled.storyName = 'with SourcegraphDotCom disabled'

export const With1AutoDefinedContext: Story = () => (
    <WebStory>{() => <SearchContextsListTab {...propsWithContexts} />}</WebStory>
)

With1AutoDefinedContext.storyName = 'with 1 auto-defined context'

export const With2AutoDefinedContexts: Story = () => (
    <WebStory>
        {() => (
            <SearchContextsListTab
                {...propsWithContexts}
                fetchAutoDefinedSearchContexts={mockFetchAutoDefinedSearchContexts(2)}
            />
        )}
    </WebStory>
)

With2AutoDefinedContexts.storyName = 'with 2 auto-defined contexts'

export const With3AutoDefinedContexts: Story = () => (
    <WebStory>
        {() => (
            <SearchContextsListTab
                {...propsWithContexts}
                fetchAutoDefinedSearchContexts={mockFetchAutoDefinedSearchContexts(3)}
            />
        )}
    </WebStory>
)

With3AutoDefinedContexts.storyName = 'with 3 auto-defined contexts'

export const With4AutoDefinedContexts: Story = () => (
    <WebStory>
        {() => (
            <SearchContextsListTab
                {...propsWithContexts}
                fetchAutoDefinedSearchContexts={mockFetchAutoDefinedSearchContexts(4)}
            />
        )}
    </WebStory>
)

With4AutoDefinedContexts.storyName = 'with 4 auto-defined contexts'
