import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import React from 'react'
import { NEVER, Observable, of, throwError } from 'rxjs'

import { IRepository, ISearchContext, ISearchContextRepositoryRevisions } from '@sourcegraph/shared/src/graphql/schema'

import { WebStory } from '../../components/WebStory'

import { SearchContextPage } from './SearchContextPage'

const { add } = storiesOf('web/searchContexts/SearchContextPage', module)
    .addParameters({
        chromatic: { viewports: [1200] },
    })
    .addDecorator(story => <div className="p-3 container">{story()}</div>)

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

add(
    'public context',
    () => (
        <WebStory>
            {webProps => <SearchContextPage {...webProps} fetchSearchContextBySpec={fetchPublicContext} />}
        </WebStory>
    ),
    {}
)

add(
    'autodefined context',
    () => (
        <WebStory>
            {webProps => <SearchContextPage {...webProps} fetchSearchContextBySpec={fetchAutoDefinedContext} />}
        </WebStory>
    ),
    {}
)

add(
    'private context',
    () => (
        <WebStory>
            {webProps => <SearchContextPage {...webProps} fetchSearchContextBySpec={fetchPrivateContext} />}
        </WebStory>
    ),
    {}
)

add(
    'loading',
    () => <WebStory>{webProps => <SearchContextPage {...webProps} fetchSearchContextBySpec={() => NEVER} />}</WebStory>,
    {}
)

add(
    'error',
    () => (
        <WebStory>
            {webProps => (
                <SearchContextPage
                    {...webProps}
                    fetchSearchContextBySpec={() => throwError(new Error('Failed to fetch search context'))}
                />
            )}
        </WebStory>
    ),
    {}
)
