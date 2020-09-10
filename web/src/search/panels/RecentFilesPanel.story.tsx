import React from 'react'
import { of, NEVER } from 'rxjs'
import { RecentFilesPanel } from './RecentFilesPanel'
import { storiesOf } from '@storybook/react'
import { WebStory } from '../../components/WebStory'
import { _fetchRecentFiles } from './utils'

const { add } = storiesOf('web/search/panels/RecentFilesPanel', module)
    .addParameters({
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
        },
        chromatic: { viewports: [800] },
    })
    .addDecorator(story => <div style={{ width: '800px' }}>{story()}</div>)

const emptyRecentFiles = {
    totalCount: 0,
    nodes: [],
    pageInfo: {
        endCursor: null,
        hasNextPage: false,
    },
}

const props = {
    authenticatedUser: null,
    fetchRecentFiles: _fetchRecentFiles,
}

add('Populated', () => <WebStory>{() => <RecentFilesPanel {...props} />}</WebStory>)

add('Loading', () => <WebStory>{() => <RecentFilesPanel {...props} fetchRecentFiles={() => NEVER} />}</WebStory>)

add('Empty', () => (
    <WebStory>{() => <RecentFilesPanel {...props} fetchRecentFiles={() => of(emptyRecentFiles)} />}</WebStory>
))
