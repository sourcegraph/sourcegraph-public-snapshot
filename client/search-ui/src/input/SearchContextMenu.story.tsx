import { storiesOf } from '@storybook/react'
import { Observable, of } from 'rxjs'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { ListSearchContextsResult } from '@sourcegraph/search'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { SearchContextMenu, SearchContextMenuProps } from './SearchContextMenu'

const { add } = storiesOf('search-ui/input/SearchContextMenu', module)
    .addParameters({
        chromatic: { viewports: [500], disableSnapshot: false },
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
    platformContext: NOOP_PLATFORM_CONTEXT,
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

const emptySearchContexts = {
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
    fetchSearchContexts: mockFetchSearchContexts,
}

add('default', () => <BrandedStory>{() => <SearchContextMenu {...defaultProps} />}</BrandedStory>, {})

add(
    'empty',
    () => <BrandedStory>{() => <SearchContextMenu {...defaultProps} {...emptySearchContexts} />}</BrandedStory>,
    {}
)

add(
    'with manage link',
    () => (
        <BrandedStory>{() => <SearchContextMenu {...defaultProps} showSearchContextManagement={true} />}</BrandedStory>
    ),
    {}
)
