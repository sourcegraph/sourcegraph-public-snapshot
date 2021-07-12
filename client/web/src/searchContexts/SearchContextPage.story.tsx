import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import React from 'react'
import { NEVER, Observable, of, throwError } from 'rxjs'

import { IRepository, ISearchContext, ISearchContextRepositoryRevisions } from '@sourcegraph/shared/src/graphql/schema'

import { WebStory } from '../components/WebStory'

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

const fetchPublicContext = (): Observable<ISearchContext> =>
    of({
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
    })

const fetchPrivateContext = (): Observable<ISearchContext> =>
    of({
        __typename: 'SearchContext',
        id: '1',
        spec: 'private-ctx',
        name: 'private-ctx',
        namespace: null,
        public: false,
        autoDefined: false,
        description: 'Repositories on Sourcegraph',
        repositories,
        updatedAt: subDays(new Date(), 1).toISOString(),
        viewerCanManage: true,
    })

add(
    'public context',
    () => (
        <WebStory>{webProps => <SearchContextPage {...webProps} fetchSearchContext={fetchPublicContext} />}</WebStory>
    ),
    {}
)

add(
    'private context',
    () => (
        <WebStory>{webProps => <SearchContextPage {...webProps} fetchSearchContext={fetchPrivateContext} />}</WebStory>
    ),
    {}
)

add(
    'loading',
    () => <WebStory>{webProps => <SearchContextPage {...webProps} fetchSearchContext={() => NEVER} />}</WebStory>,
    {}
)

add(
    'error',
    () => (
        <WebStory>
            {webProps => (
                <SearchContextPage
                    {...webProps}
                    fetchSearchContext={() => throwError(new Error('Failed to fetch search context'))}
                />
            )}
        </WebStory>
    ),
    {}
)
