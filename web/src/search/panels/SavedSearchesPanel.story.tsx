import React from 'react'
import { storiesOf } from '@storybook/react'
import { WebStory } from '../../components/WebStory'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'
import { authUser, _fetchSavedSearches } from './utils'
import { SavedSearchesPanel } from './SavedSearchesPanel'
import { NEVER, of } from 'rxjs'

const { add } = storiesOf('web/search/panels/SavedSearchesPanel', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
    },
})

const props = {
    authenticatedUser: authUser,
    patternType: SearchPatternType.literal,
    fetchSavedSearches: _fetchSavedSearches,
}

add('Populated', () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <SavedSearchesPanel {...props} />
            </div>
        )}
    </WebStory>
))

add('Loading', () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <SavedSearchesPanel {...props} fetchSavedSearches={() => NEVER} />
            </div>
        )}
    </WebStory>
))

add('Empty', () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <SavedSearchesPanel {...props} fetchSavedSearches={() => of([])} />
            </div>
        )}
    </WebStory>
))
