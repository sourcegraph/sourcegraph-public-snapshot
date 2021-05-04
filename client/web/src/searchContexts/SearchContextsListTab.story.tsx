import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import { createMemoryHistory } from 'history'
import React from 'react'
import { Observable, of } from 'rxjs'

import { WebStory } from '../components/WebStory'
import { ListSearchContextsResult, Scalars, SearchContextsNamespaceFilterType } from '../graphql-operations'

import { SearchContextsListTab, SearchContextsListTabProps } from './SearchContextsListTab'
import { mockFetchAutoDefinedSearchContexts, mockFetchSearchContexts } from './testHelpers'

const { add } = storiesOf('web/searchContexts/SearchContextsListTab', module)
    .addParameters({
        chromatic: { viewports: [1200] },
    })
    .addDecorator(story => (
        <div className="dropdown-menu show" style={{ position: 'static' }}>
            {story()}
        </div>
    ))

const history = createMemoryHistory()
const defaultProps: SearchContextsListTabProps = {
    history,
    location: history.location,
    authenticatedUser: null,
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
    fetchSearchContexts: mockFetchSearchContexts,
}

const propsWithContexts: SearchContextsListTabProps = {
    ...defaultProps,
    fetchAutoDefinedSearchContexts: of([
        {
            __typename: 'SearchContext',
            id: '1',
            spec: 'global',
            autoDefined: true,
            description: 'All repositories on Sourcegraph',
            repositories: [],
            updatedAt: subDays(new Date(), 1).toISOString(),
        },
        {
            __typename: 'SearchContext',
            id: '2',
            spec: '@username',
            autoDefined: true,
            description: 'Your repositories on Sourcegraph',
            repositories: [],
            updatedAt: subDays(new Date(), 1).toISOString(),
        },
    ]),
    fetchSearchContexts: ({
        first,
        namespaceFilterType,
        namespace,
        query,
        after,
    }: {
        first: number
        query?: string
        namespace?: Scalars['ID']
        namespaceFilterType?: SearchContextsNamespaceFilterType
        after?: string
    }): Observable<ListSearchContextsResult['searchContexts']> =>
        of({
            nodes: [
                {
                    __typename: 'SearchContext',
                    id: '3',
                    spec: '@username/test-version-1.5',
                    autoDefined: false,
                    description: 'Only code in version 1.5',
                    updatedAt: subDays(new Date(), 1).toISOString(),
                    repositories: [],
                },
            ],
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            totalCount: 1,
        }),
}

add('default', () => <WebStory>{() => <SearchContextsListTab {...defaultProps} />}</WebStory>, {})

add('with contexts', () => <WebStory>{() => <SearchContextsListTab {...propsWithContexts} />}</WebStory>, {})
