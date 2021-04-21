import { storiesOf } from '@storybook/react'
import React from 'react'
import { Observable, of } from 'rxjs'

import { Scalars, SearchContextsNamespaceFilterType } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../components/WebStory'
import { ListSearchContextsResult } from '../../graphql-operations'
import { mockFetchAutoDefinedSearchContexts, mockFetchSearchContexts } from '../../searchContexts/testHelpers'

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
    showSearchContextManagement: false,
    fetchAutoDefinedSearchContexts: of([
        {
            __typename: 'SearchContext',
            id: '1',
            spec: 'global',
            autoDefined: true,
            description: 'All repositories on Sourcegraph',
            repositories: [],
        },
        {
            __typename: 'SearchContext',
            id: '2',
            spec: '@username',
            autoDefined: true,
            description: 'Your repositories on Sourcegraph',
            repositories: [],
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
