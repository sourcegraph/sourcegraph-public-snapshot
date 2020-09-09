import React from 'react'
import { EnterpriseHomePanels } from './EnterpriseHomePanels'
import { storiesOf } from '@storybook/react'
import { WebStory } from '../../components/WebStory'
import {
    IOrg,
    SearchPatternType,
    ISavedSearch,
    Namespace,
    IUser,
    IUserConnection,
    IConfiguration,
} from '../../../../shared/src/graphql/schema'
import { AuthenticatedUser } from '../../auth'
import { of, Observable } from 'rxjs'
import { NOOP_SETTINGS_CASCADE } from '../../../../shared/src/util/searchTestHelpers'

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

const org: IOrg = {
    __typename: 'Org',
    id: '1',
    name: 'test-org',
    displayName: 'test org',
    createdAt: '2020-01-01',
    members: {
        __typename: 'UserConnection',
        nodes: [authUser as IUser],
        totalCount: 1,
        pageInfo: { __typename: 'PageInfo', endCursor: null, hasNextPage: false },
    },
    latestSettings: null,
    settingsCascade: {
        __typename: 'SettingsCascade',
        subjects: [],
        final: '',
        merged: { __typename: 'Configuration', contents: '', messages: [] },
    },
    configurationCascade: {
        __typename: 'ConfigurationCascade',
        subjects: [],
        merged: { __typename: 'Configuration', contents: '', messages: [] },
    },
    viewerPendingInvitation: null,
    viewerCanAdminister: true,
    viewerIsMember: true,
    url: '/organizations/test-org',
    settingsURL: '/organizations/test-org/settings',
    namespaceName: 'test-org',
    campaigns: {
        __typename: 'CampaignConnection',
        nodes: [],
        totalCount: 0,
        pageInfo: { __typename: 'PageInfo', endCursor: null, hasNextPage: false },
    },
}

const props = {
    authenticatedUser: authUser,
    patternType: SearchPatternType.literal,
    fetchSavedSearches: (): Observable<ISavedSearch[]> =>
        of([
            {
                __typename: 'SavedSearch',
                id: 'test',
                description: 'test',
                query: 'test',
                notify: false,
                notifySlack: false,
                namespace: authUser as Namespace,
                slackWebhookURL: null,
            },
            {
                __typename: 'SavedSearch',
                id: 'test-org',
                description: 'org test',
                query: 'org test',
                notify: false,
                notifySlack: false,
                namespace: org,
                slackWebhookURL: null,
            },
        ]),
}

add('Panels', () => <WebStory>{() => <EnterpriseHomePanels {...props} />}</WebStory>)
