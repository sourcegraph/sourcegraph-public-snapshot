import { storiesOf } from '@storybook/react'
import React from 'react'
import { Observable, of } from 'rxjs'

import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'

import { WebStory } from '../../components/WebStory'
import { ListSearchContextsResult } from '../../graphql-operations'

import { SearchContextMenu, SearchContextMenuProps } from './SearchContextMenu'

const { add } = storiesOf('web/search/input/SearchContextMenu', module)
    .addParameters({
        chromatic: { viewports: [500] },
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/4Fy9rURbfF2bsl4BvYunUO/RFC-261-Search-Contexts?node-id=581%3A4754',
        },
    })
    .addDecorator(story => (
        <div className="dropdown-menu show" style={{ position: 'static' }}>
            {story()}
        </div>
    ))

const defaultProps: SearchContextMenuProps = {
    authenticatedUser: null,
    showSearchContextManagement: false,
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(2),
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
                    updatedAt: '2021-03-15T19:39:11Z',
                    viewerCanManage: true,
                    query: '',
                    repositories: [],
                },
            ],
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            totalCount: 1,
        }),
    defaultSearchContextSpec: 'global',
    selectedSearchContextSpec: 'global',
    selectSearchContextSpec: () => {},
    closeMenu: () => {},
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    searchContextsEnabled: true,
}

const emptySearchContexts = {
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
    fetchSearchContexts: mockFetchSearchContexts,
}

add('default', () => <WebStory>{() => <SearchContextMenu {...defaultProps} />}</WebStory>, {})

add('empty', () => <WebStory>{() => <SearchContextMenu {...defaultProps} {...emptySearchContexts} />}</WebStory>, {})

add(
    'with manage link',
    () => <WebStory>{() => <SearchContextMenu {...defaultProps} showSearchContextManagement={true} />}</WebStory>,
    {}
)
