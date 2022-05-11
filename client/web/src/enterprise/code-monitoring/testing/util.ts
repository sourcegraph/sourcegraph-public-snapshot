import { AuthenticatedUser } from '../../../auth'
import {
    CodeMonitorFields,
    EventStatus,
    ListCodeMonitors,
    MonitorTriggerEventsResult,
} from '../../../graphql-operations'

export const mockUser: AuthenticatedUser = {
    __typename: 'User',
    id: 'userID',
    username: 'username',
    email: 'user@me.com',
    siteAdmin: true,
    databaseID: 0,
    tags: [],
    url: '',
    avatarURL: '',
    displayName: 'display name',
    settingsURL: '',
    viewerCanAdminister: true,
    organizations: {
        __typename: 'OrgConnection',
        nodes: [],
    },
    session: { __typename: 'Session', canSignOut: true },
    tosAccepted: true,
    searchable: true,
}

export const mockCodeMonitorFields: CodeMonitorFields = {
    __typename: 'Monitor',
    id: 'foo0',
    description: 'Test code monitor',
    enabled: true,
    trigger: { id: 'test-0', query: 'test' },
    actions: {
        nodes: [
            {
                __typename: 'MonitorEmail',
                id: 'test-action-0',
                enabled: true,
                includeResults: false,
                recipients: { nodes: [{ id: 'baz-0' }] },
            },
        ],
    },
}

export const mockCodeMonitor = {
    node: {
        __typename: 'Monitor',
        id: 'foo0',
        description: 'Test code monitor',
        enabled: true,
        owner: { id: 'test-id', namespaceName: 'test-user' },
        actions: {
            nodes: [
                {
                    __typename: 'MonitorEmail',
                    id: 'test-action-0',
                    enabled: true,
                    includeResults: false,
                    recipients: { nodes: [{ id: 'baz-0', url: '/user/test' }] },
                },
                {
                    __typename: 'MonitorSlackWebhook',
                    id: 'test-action-1',
                    enabled: true,
                    includeResults: false,
                    url: 'https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX',
                },
            ],
        },
        trigger: { id: 'test-0', query: 'test' },
    },
}

export const mockCodeMonitorNodes: ListCodeMonitors['nodes'] = [
    {
        id: 'foo0',
        description: 'Test code monitor',
        enabled: true,
        actions: {
            nodes: [
                {
                    __typename: 'MonitorEmail',
                    id: 'test-action-0 ',
                    enabled: true,
                    includeResults: false,
                    recipients: { nodes: [{ id: 'baz-0' }] },
                },
            ],
        },
        trigger: { id: 'test-0', query: 'test' },
    },
    {
        id: 'foo1',
        description: 'Second test code monitor',
        enabled: true,
        actions: {
            nodes: [
                {
                    __typename: 'MonitorSlackWebhook',
                    id: 'test-action-1 ',
                    enabled: true,
                    includeResults: false,
                    url: 'https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX',
                },
            ],
        },
        trigger: { id: 'test-1', query: 'test' },
    },
    {
        id: 'foo2',
        description: 'Third test code monitor',
        enabled: true,
        actions: {
            nodes: [
                {
                    __typename: 'MonitorEmail',
                    id: 'test-action-2 ',
                    enabled: true,
                    includeResults: false,
                    recipients: { nodes: [{ id: 'baz-2' }] },
                },
                {
                    __typename: 'MonitorSlackWebhook',
                    id: 'test-action-1 ',
                    enabled: true,
                    includeResults: false,
                    url: 'https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX',
                },
            ],
        },
        trigger: { id: 'test-2', query: 'test' },
    },
    {
        id: 'foo3',
        description: 'Fourth test code monitor',
        enabled: true,
        actions: {
            nodes: [
                {
                    __typename: 'MonitorEmail',
                    id: 'test-action-3 ',
                    enabled: true,
                    includeResults: false,
                    recipients: { nodes: [{ id: 'baz-3' }] },
                },
                {
                    __typename: 'MonitorWebhook',
                    id: 'test-action-4',
                    enabled: true,
                    includeResults: false,
                    url: 'https://example.com/webhook',
                },
            ],
        },
        trigger: { id: 'test-3', query: 'test' },
    },
    {
        id: 'foo4',
        description: 'Fifth test code monitor',
        enabled: true,
        actions: {
            nodes: [
                {
                    __typename: 'MonitorWebhook',
                    id: 'test-action-4',
                    enabled: true,
                    includeResults: false,
                    url: 'https://example.com/webhook',
                },
            ],
        },
        trigger: { id: 'test-4', query: 'test' },
    },
    {
        id: 'foo5',
        description: 'Sixth test code monitor',
        enabled: true,
        actions: {
            nodes: [
                {
                    __typename: 'MonitorEmail',
                    id: 'test-action-5 ',
                    enabled: true,
                    includeResults: false,
                    recipients: { nodes: [{ id: 'baz-5' }] },
                },
            ],
        },
        trigger: { id: 'test-5', query: 'test' },
    },
    {
        id: 'foo6',
        description: 'Seventh test code monitor',
        enabled: true,
        actions: {
            nodes: [
                {
                    __typename: 'MonitorEmail',
                    id: 'test-action-6 ',
                    enabled: true,
                    includeResults: false,
                    recipients: { nodes: [{ id: 'baz-6' }] },
                },
            ],
        },
        trigger: { id: 'test-6', query: 'test' },
    },
    {
        id: 'foo7',
        description: 'Eighth test code monitor',
        enabled: true,
        actions: {
            nodes: [
                {
                    __typename: 'MonitorEmail',
                    id: 'test-action-7 ',
                    enabled: true,
                    includeResults: false,
                    recipients: { nodes: [{ id: 'baz-7' }] },
                },
            ],
        },
        trigger: { id: 'test-7', query: 'test' },
    },
    {
        id: 'foo9',
        description: 'Ninth test code monitor',
        enabled: true,
        actions: {
            nodes: [
                {
                    __typename: 'MonitorEmail',
                    id: 'test-action-9 ',
                    enabled: true,
                    includeResults: false,
                    recipients: { nodes: [{ id: 'baz-9' }] },
                },
            ],
        },
        trigger: { id: 'test-9', query: 'test' },
    },
    {
        id: 'foo10',
        description: 'Tenth test code monitor',
        enabled: true,
        actions: {
            nodes: [
                {
                    __typename: 'MonitorEmail',
                    id: 'test-action-0 ',
                    enabled: true,
                    includeResults: false,
                    recipients: { nodes: [{ id: 'baz-0' }] },
                },
            ],
        },
        trigger: { id: 'test-0', query: 'test' },
    },
    {
        id: 'foo11',
        description: 'Eleventh test code monitor',
        enabled: true,
        actions: {
            nodes: [
                {
                    __typename: 'MonitorEmail',
                    id: 'test-action-1 ',
                    enabled: true,
                    includeResults: false,
                    recipients: { nodes: [{ id: 'baz-1' }] },
                },
            ],
        },
        trigger: { id: 'test-1', query: 'test' },
    },
    {
        id: 'foo12',
        description: 'Twelfth test code monitor',
        enabled: true,
        actions: {
            nodes: [
                {
                    __typename: 'MonitorEmail',
                    id: 'test-action-2 ',
                    enabled: true,
                    includeResults: false,
                    recipients: { nodes: [{ id: 'baz-2' }] },
                },
            ],
        },
        trigger: { id: 'test-2', query: 'test' },
    },
]

// Only minimal authenticated user data is needed for the code monitor tests
// eslint-disable-next-line @typescript-eslint/consistent-type-assertions
export const mockAuthenticatedUser: AuthenticatedUser = {
    id: 'userID',
    username: 'username',
    email: 'user@me.com',
    siteAdmin: true,
} as AuthenticatedUser

export const mockLogs: MonitorTriggerEventsResult = {
    currentUser: {
        monitors: {
            __typename: 'MonitorConnection',
            nodes: [
                {
                    __typename: 'Monitor',
                    description: 'First test code monitor',
                    id: '123',
                    trigger: {
                        __typename: 'MonitorQuery',
                        query: 'test type:diff',
                        events: {
                            __typename: 'MonitorTriggerEventConnection',
                            nodes: [
                                {
                                    __typename: 'MonitorTriggerEvent',
                                    id: 'a',
                                    status: EventStatus.SUCCESS,
                                    message: null,
                                    timestamp: '2022-02-11T18:30:50Z',
                                    query: 'test type:diff',
                                    resultCount: 1,
                                    actions: {
                                        __typename: 'MonitorActionConnection',
                                        nodes: [
                                            {
                                                __typename: 'MonitorEmail',
                                                events: {
                                                    __typename: 'MonitorActionEventConnection',
                                                    nodes: [],
                                                },
                                            },
                                            {
                                                __typename: 'MonitorSlackWebhook',
                                                events: {
                                                    __typename: 'MonitorActionEventConnection',
                                                    nodes: [
                                                        {
                                                            id: 'ad',
                                                            __typename: 'MonitorActionEvent',
                                                            status: EventStatus.SUCCESS,
                                                            message: null,
                                                            timestamp: '2022-02-11T18:30:54Z',
                                                        },
                                                    ],
                                                },
                                            },
                                        ],
                                    },
                                },
                                {
                                    __typename: 'MonitorTriggerEvent',
                                    id: 'b',
                                    status: EventStatus.SUCCESS,
                                    message: null,
                                    timestamp: '2022-02-11T18:24:49Z',
                                    query: 'test type:diff',
                                    resultCount: 2,
                                    actions: {
                                        __typename: 'MonitorActionConnection',
                                        nodes: [
                                            {
                                                __typename: 'MonitorEmail',
                                                events: {
                                                    __typename: 'MonitorActionEventConnection',
                                                    nodes: [],
                                                },
                                            },
                                            {
                                                __typename: 'MonitorSlackWebhook',
                                                events: {
                                                    __typename: 'MonitorActionEventConnection',
                                                    nodes: [
                                                        {
                                                            id: 'ad',
                                                            __typename: 'MonitorActionEvent',
                                                            status: EventStatus.SUCCESS,
                                                            message: null,
                                                            timestamp: '2022-02-11T18:24:53Z',
                                                        },
                                                    ],
                                                },
                                            },
                                        ],
                                    },
                                },
                                {
                                    __typename: 'MonitorTriggerEvent',
                                    id: 'c',
                                    status: EventStatus.SUCCESS,
                                    message: null,
                                    timestamp: '2022-02-11T17:29:16Z',
                                    query: 'test type:diff',
                                    resultCount: 5,
                                    actions: {
                                        __typename: 'MonitorActionConnection',
                                        nodes: [
                                            {
                                                __typename: 'MonitorEmail',
                                                events: {
                                                    __typename: 'MonitorActionEventConnection',
                                                    nodes: [],
                                                },
                                            },
                                            {
                                                __typename: 'MonitorSlackWebhook',
                                                events: {
                                                    __typename: 'MonitorActionEventConnection',
                                                    nodes: [
                                                        {
                                                            id: 'ac',
                                                            __typename: 'MonitorActionEvent',
                                                            status: EventStatus.SUCCESS,
                                                            message: null,
                                                            timestamp: '2022-02-11T17:29:21Z',
                                                        },
                                                    ],
                                                },
                                            },
                                        ],
                                    },
                                },
                            ],
                            totalCount: 60,
                            pageInfo: {
                                endCursor: 'c',
                                hasNextPage: true,
                            },
                        },
                    },
                },
                {
                    __typename: 'Monitor',
                    description: 'Second test code monitor (no events)',
                    id: '456',
                    trigger: {
                        __typename: 'MonitorQuery',
                        query: 'test type:commit',
                        events: {
                            __typename: 'MonitorTriggerEventConnection',
                            nodes: [],
                            totalCount: 0,
                            pageInfo: { endCursor: '', hasNextPage: false },
                        },
                    },
                },
                {
                    __typename: 'Monitor',
                    description: 'Third test code monitor (error in query)',
                    id: '789',
                    trigger: {
                        __typename: 'MonitorQuery',
                        query: 'test type:commit',
                        events: {
                            __typename: 'MonitorTriggerEventConnection',
                            nodes: [
                                {
                                    __typename: 'MonitorTriggerEvent',
                                    id: 'd',
                                    status: EventStatus.ERROR,
                                    message:
                                        'Search failed. This is a very long error that should wrap. Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.',
                                    timestamp: '2022-02-14T12:29:21Z',
                                    query: '',
                                    resultCount: 0,
                                    actions: {
                                        __typename: 'MonitorActionConnection',
                                        nodes: [
                                            {
                                                __typename: 'MonitorSlackWebhook',
                                                events: {
                                                    __typename: 'MonitorActionEventConnection',
                                                    nodes: [
                                                        {
                                                            id: 'ab',
                                                            __typename: 'MonitorActionEvent',
                                                            status: EventStatus.SUCCESS,
                                                            message: null,
                                                            timestamp: '2022-02-14T12:29:21Z',
                                                        },
                                                    ],
                                                },
                                            },
                                        ],
                                    },
                                },
                            ],
                            totalCount: 1,
                            pageInfo: { endCursor: '', hasNextPage: false },
                        },
                    },
                },
                {
                    __typename: 'Monitor',
                    description: 'Fourt test code monitor (error in action)',
                    id: '101112',
                    trigger: {
                        __typename: 'MonitorQuery',
                        query: 'test type:commit',
                        events: {
                            __typename: 'MonitorTriggerEventConnection',
                            nodes: [
                                {
                                    __typename: 'MonitorTriggerEvent',
                                    id: 'e',
                                    status: EventStatus.SUCCESS,
                                    message: null,
                                    timestamp: '2022-02-13T12:29:21Z',
                                    query: '',
                                    resultCount: 0,
                                    actions: {
                                        __typename: 'MonitorActionConnection',
                                        nodes: [
                                            {
                                                __typename: 'MonitorEmail',
                                                events: {
                                                    __typename: 'MonitorActionEventConnection',
                                                    nodes: [],
                                                },
                                            },
                                            {
                                                __typename: 'MonitorSlackWebhook',
                                                events: {
                                                    __typename: 'MonitorActionEventConnection',
                                                    nodes: [
                                                        {
                                                            __typename: 'MonitorActionEvent',
                                                            id: 'aa',
                                                            status: EventStatus.ERROR,
                                                            message: 'Calling webhook failed',
                                                            timestamp: '2022-02-13T12:29:21Z',
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                __typename: 'MonitorWebhook',
                                                events: {
                                                    __typename: 'MonitorActionEventConnection',
                                                    nodes: [],
                                                },
                                            },
                                        ],
                                    },
                                },
                            ],
                            totalCount: 1,
                            pageInfo: { endCursor: '', hasNextPage: false },
                        },
                    },
                },
            ],
            pageInfo: {
                endCursor: '123',
                hasNextPage: false,
            },
            totalCount: 2,
        },
    },
}
