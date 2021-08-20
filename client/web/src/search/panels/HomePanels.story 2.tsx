import { storiesOf } from '@storybook/react'
import { parseISO } from 'date-fns'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../components/WebStory'
import { SearchPatternType } from '../../graphql-operations'

import { HomePanels } from './HomePanels'
import { _fetchRecentFileViews, _fetchRecentSearches, _fetchSavedSearches, authUser } from './utils'

const { add } = storiesOf('web/search/panels/HomePanels', module).addParameters({
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
    telemetryService: NOOP_TELEMETRY_SERVICE,
    showEnterpriseHomePanels: true,
    isSourcegraphDotCom: false,
}

add('Panels on Server', () => <WebStory>{() => <HomePanels {...props} />}</WebStory>)

add('Panels on Cloud', () => <WebStory>{() => <HomePanels {...props} isSourcegraphDotCom={true} />}</WebStory>)
