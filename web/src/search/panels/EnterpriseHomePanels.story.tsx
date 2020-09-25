import React from 'react'
import { _fetchRecentFileViews, _fetchRecentSearches, _fetchSavedSearches, authUser } from './utils'
import { EnterpriseHomePanels } from './EnterpriseHomePanels'
import { parseISO } from 'date-fns'
import { SearchPatternType } from '../../graphql-operations'
import { storiesOf } from '@storybook/react'
import { WebStory } from '../../components/WebStory'

const { add } = storiesOf('web/search/panels/EnterpriseHomePanels', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
    },
    chromatic: { viewports: [480, 1200] },
})

const props = {
    authenticatedUser: authUser,
    patternType: SearchPatternType.literal,
    fetchSavedSearches: _fetchSavedSearches,
    fetchRecentSearches: _fetchRecentSearches,
    fetchRecentFileViews: _fetchRecentFileViews,
    now: () => parseISO('2020-09-16T23:15:01Z'),
}

add('Panels', () => <WebStory>{() => <EnterpriseHomePanels {...props} />}</WebStory>)
