import React from 'react'
import { of, NEVER } from 'rxjs'
import { RecentSearchesPanel } from './RecentSearchesPanel'
import { storiesOf } from '@storybook/react'
import { WebStory } from '../../components/WebStory'
import { _fetchRecentSearches } from './utils'

const { add } = storiesOf('web/search/panels/RecentSearchesPanel', module)
    .addParameters({
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
        },
        chromatic: { viewports: [800] },
    })
    .addDecorator(story => <div style={{ width: '800px' }}>{story()}</div>)

const emptyRecentSearches = {
    totalCount: 0,
    nodes: [],
    pageInfo: {
        endCursor: null,
        hasNextPage: false,
    },
}

const props = {
    authenticatedUser: null,
    fetchRecentSearches: _fetchRecentSearches,
}

add('Populated', () => <WebStory>{() => <RecentSearchesPanel {...props} />}</WebStory>)

add('Loading', () => <WebStory>{() => <RecentSearchesPanel {...props} fetchRecentSearches={() => NEVER} />}</WebStory>)

add('Empty', () => (
    <WebStory>{() => <RecentSearchesPanel {...props} fetchRecentSearches={() => of(emptyRecentSearches)} />}</WebStory>
))
