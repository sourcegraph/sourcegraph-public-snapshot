import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import React from 'react'
import { Observable, of } from 'rxjs'

import { ISearchContext } from '@sourcegraph/shared/src/graphql/schema'

import { WebStory } from '../components/WebStory'

import { SearchContextPage } from './SearchContextPage'

const { add } = storiesOf('web/searchContexts/SearchContextPage', module)
    .addParameters({
        chromatic: { viewports: [1200] },
    })
    .addDecorator(story => (
        <div className="p-3 container web-content" style={{ position: 'static' }}>
            {story()}
        </div>
    ))

const fetchPublicContext = (): Observable<ISearchContext> =>
    of({
        __typename: 'SearchContext',
        id: '1',
        spec: 'public-ctx',
        public: true,
        autoDefined: true,
        description: 'Repositories on Sourcegraph',
        repositories: [],
        updatedAt: subDays(new Date(), 1).toISOString(),
    })

const fetchPrivateContext = (): Observable<ISearchContext> =>
    of({
        __typename: 'SearchContext',
        id: '1',
        spec: 'private-ctx',
        public: false,
        autoDefined: true,
        description: 'Repositories on Sourcegraph',
        repositories: [],
        updatedAt: subDays(new Date(), 1).toISOString(),
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
