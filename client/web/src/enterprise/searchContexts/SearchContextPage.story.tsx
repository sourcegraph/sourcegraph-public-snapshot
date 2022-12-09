import { DecoratorFn, Meta, Story } from '@storybook/react'
import { subDays } from 'date-fns'
import { NEVER, Observable, of, throwError } from 'rxjs'

import { SearchContextFields, SearchContextRepositoryRevisionsFields } from '@sourcegraph/search'
import { mockAuthenticatedUser } from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { WebStory } from '../../components/WebStory'

import { SearchContextPage } from './SearchContextPage'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/enterprise/searchContexts/SearchContextPage',
    decorators: [decorator],
    parameters: {
        chromatic: { viewports: [1200], disableSnapshot: false },
    },
}

export default config

const repositories: SearchContextRepositoryRevisionsFields[] = [
    {
        __typename: 'SearchContextRepositoryRevisions',
        repository: {
            __typename: 'Repository',
            name: 'github.com/example/example',
        },
        revisions: ['REVISION1', 'REVISION2'],
    },
    {
        __typename: 'SearchContextRepositoryRevisions',
        repository: {
            __typename: 'Repository',
            name: 'github.com/example/really-really-really-really-really-really-long-name',
        },
        revisions: ['REVISION3', 'LONG-LONG-LONG-LONG-LONG-LONG-LONG-LONG-REVISION'],
    },
]

const mockContext: SearchContextFields = {
    __typename: 'SearchContext',
    id: '1',
    spec: 'public-ctx',
    name: 'public-ctx',
    namespace: null,
    public: true,
    autoDefined: false,
    description: 'Repositories on Sourcegraph',
    query: '',
    repositories,
    updatedAt: subDays(new Date(), 1).toISOString(),
    viewerCanManage: true,
    viewerHasAsDefault: false,
    viewerHasStarred: true,
}

const fetchPublicContext = (): Observable<SearchContextFields> => of(mockContext)

const fetchPrivateContext = (): Observable<SearchContextFields> =>
    of({
        ...mockContext,
        spec: 'private-ctx',
        name: 'private-ctx',
        namespace: null,
        public: false,
        viewerHasStarred: false,
    })

const fetchAutoDefinedContext = (): Observable<SearchContextFields> =>
    of({
        ...mockContext,
        autoDefined: true,
        viewerHasStarred: false,
        viewerHasAsDefault: true,
        spec: 'auto-ctx',
        name: 'auto-ctx',
    })

export const PublicContext: Story = () => (
    <WebStory>
        {webProps => (
            <SearchContextPage
                {...webProps}
                fetchSearchContextBySpec={fetchPublicContext}
                platformContext={NOOP_PLATFORM_CONTEXT}
                authenticatedUser={mockAuthenticatedUser}
            />
        )}
    </WebStory>
)

PublicContext.storyName = 'public context'

export const PublicContextUnauthenticated: Story = () => (
    <WebStory>
        {webProps => (
            <SearchContextPage
                {...webProps}
                fetchSearchContextBySpec={fetchPublicContext}
                platformContext={NOOP_PLATFORM_CONTEXT}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

PublicContextUnauthenticated.storyName = 'public context, unauthenticated user'

export const AutodefinedContext: Story = () => (
    <WebStory>
        {webProps => (
            <SearchContextPage
                {...webProps}
                fetchSearchContextBySpec={fetchAutoDefinedContext}
                platformContext={NOOP_PLATFORM_CONTEXT}
                authenticatedUser={mockAuthenticatedUser}
            />
        )}
    </WebStory>
)

AutodefinedContext.storyName = 'autodefined context'

export const PrivateContext: Story = () => (
    <WebStory>
        {webProps => (
            <SearchContextPage
                {...webProps}
                fetchSearchContextBySpec={fetchPrivateContext}
                platformContext={NOOP_PLATFORM_CONTEXT}
                authenticatedUser={mockAuthenticatedUser}
            />
        )}
    </WebStory>
)

PrivateContext.storyName = 'private context'

export const Loading: Story = () => (
    <WebStory>
        {webProps => (
            <SearchContextPage
                {...webProps}
                fetchSearchContextBySpec={() => NEVER}
                platformContext={NOOP_PLATFORM_CONTEXT}
                authenticatedUser={mockAuthenticatedUser}
            />
        )}
    </WebStory>
)

Loading.storyName = 'loading'

export const ErrorStory: Story = () => (
    <WebStory>
        {webProps => (
            <SearchContextPage
                {...webProps}
                fetchSearchContextBySpec={() => throwError(new Error('Failed to fetch search context'))}
                platformContext={NOOP_PLATFORM_CONTEXT}
                authenticatedUser={mockAuthenticatedUser}
            />
        )}
    </WebStory>
)

ErrorStory.storyName = 'error'
