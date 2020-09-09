import { Observable, of } from 'rxjs'
import { ISavedSearch, Namespace, IOrg } from '../../../../shared/src/graphql/schema'
import { AuthenticatedUser } from '../../auth'

export const authUser: AuthenticatedUser = {
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

export const org: IOrg = {
    __typename: 'Org',
    id: '1',
    name: 'test-org',
    displayName: 'test org',
    createdAt: '2020-01-01',
    members: {
        __typename: 'UserConnection',
        nodes: [authUser],
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

export const _fetchSavedSearches = (): Observable<ISavedSearch[]> =>
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
    ])
