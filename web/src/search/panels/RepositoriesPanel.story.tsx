import React from 'react'
import { storiesOf } from '@storybook/react'
import { _fetchRecentSearches } from './utils'
import { WebStory } from '../../components/WebStory'
import { RepositoriesPanel } from './RepositoriesPanel'
import { NEVER, of } from 'rxjs'

const { add } = storiesOf('web/search/panels/RepositoriesPanel', module)
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

add('Populated', () => <WebStory>{() => <RepositoriesPanel {...props} />}</WebStory>)

add('Loading', () => <WebStory>{() => <RepositoriesPanel {...props} fetchRecentSearches={() => NEVER} />}</WebStory>)

add('Empty', () => (
    <WebStory>{() => <RepositoriesPanel {...props} fetchRecentSearches={() => of(emptyRecentSearches)} />}</WebStory>
))
