import React from 'react'
import { EnterpriseHomePanels } from './EnterpriseHomePanels'
import { storiesOf } from '@storybook/react'
import { WebStory } from '../../components/WebStory'
import { IOrg, SearchPatternType } from '../../../../shared/src/graphql/schema'
import { AuthenticatedUser } from '../../auth'
import sinon from 'sinon'

const { add } = storiesOf('web/search/panels/EnterpriseHomePanels', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
    },
    chromatic: { viewports: [480, 769, 993, 1200] },
})

const authUser: AuthenticatedUser = {
    __typename: 'User',
    id: '0',
    email: 'alice@sourcegraph.com',
    username: 'alice',
    avatarURL: null,
    session: { canSignOut: true },
    displayName: null,
    url: '',
    settingsURL: '#',
    siteAdmin: true,
    organizations: {
        nodes: [
            { id: '0', settingsURL: '#', displayName: 'Acme Corp' },
            { id: '1', settingsURL: '#', displayName: 'Beta Inc' },
        ] as IOrg[],
    },
    tags: [],
    viewerCanAdminister: true,
    databaseID: 0,
}

const props = {
    authenticatedUser: authUser,
    patternType: SearchPatternType.literal,
    fetchSavedSearches: sinon.spy(),
}

add('Panels', () => <WebStory>{() => <EnterpriseHomePanels {...props} />}</WebStory>)
