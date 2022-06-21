import { DecoratorFn, Meta, Story } from '@storybook/react'
import { subDays } from 'date-fns'
import { NEVER, Observable, of, throwError } from 'rxjs'

import { IRepository, ISearchContext, ISearchContextRepositoryRevisions } from '@sourcegraph/shared/src/schema'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { WebStory } from '../../components/WebStory'

import { SearchContextPage } from './SearchContextPage'

const repositories: ISearchContextRepositoryRevisions[] = [
    {
        __typename: 'SearchContextRepositoryRevisions',
        repository: {
            __typename: 'Repository',
            name: 'github.com/example/example',
        } as IRepository,
        revisions: ['REVISION1', 'REVISION2'],
    },
    {
        __typename: 'SearchContextRepositoryRevisions',
        repository: {
            __typename: 'Repository',
            name: 'github.com/example/really-really-really-really-really-really-long-name',
        } as IRepository,
        revisions: ['REVISION3', 'LONG-LONG-LONG-LONG-LONG-LONG-LONG-LONG-REVISION'],
    },
]

const mockContext: ISearchContext = {
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
}

const fetchPublicContext = (): Observable<ISearchContext> => of(mockContext)

const fetchPrivateContext = (): Observable<ISearchContext> =>
    of({
        ...mockContext,
        spec: 'private-ctx',
        name: 'private-ctx',
        namespace: null,
        public: false,
    })

const fetchAutoDefinedContext = (): Observable<ISearchContext> =>
    of({
        ...mockContext,
        autoDefined: true,
    })

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/enterprise/searchContexts/SearchContextPage',
    decorators: [decorator],
    parameters: {
        chromatic: { viewports: [1200], disableSnapshot: false },
    },
}

export default config

export const PublicContext: Story = () => (
    <WebStory>
        {webProps => (
            <SearchContextPage
                {...webProps}
                fetchSearchContextBySpec={fetchPublicContext}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )}
    </WebStory>
)

PublicContext.storyName = 'public context'

export const AutodefinedContext: Story = () => (
    <WebStory>
        {webProps => (
            <SearchContextPage
                {...webProps}
                fetchSearchContextBySpec={fetchAutoDefinedContext}
                platformContext={NOOP_PLATFORM_CONTEXT}
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
            />
        )}
    </WebStory>
)

Loading.storyName = 'loading'

export const _Error: Story = () => (
    <WebStory>
        {webProps => (
            <SearchContextPage
                {...webProps}
                fetchSearchContextBySpec={() => throwError(new Error('Failed to fetch search context'))}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )}
    </WebStory>
)

_Error.storyName = 'error'
