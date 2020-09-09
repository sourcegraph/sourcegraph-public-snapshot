import React from 'react'
import { EnterpriseHomePanels } from './EnterpriseHomePanels'
import { storiesOf } from '@storybook/react'
import { WebStory } from '../../components/WebStory'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'
import { authUser, _fetchSavedSearches } from './utils'

const { add } = storiesOf('web/search/panels/EnterpriseHomePanels', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
    },
    chromatic: { viewports: [480, 769, 993, 1200] },
})

const props = {
    authenticatedUser: authUser,
    patternType: SearchPatternType.literal,
    fetchSavedSearches: _fetchSavedSearches,
}

add('Panels', () => <WebStory>{() => <EnterpriseHomePanels {...props} />}</WebStory>)
