export const mockCodeMonitor = {
    node: {
        id: 'foo0',
        description: 'Test code monitor',
        enabled: true,
        owner: { id: 'test-id', namespaceName: 'test-user' },
        actions: {
            id: 'test-0',
            enabled: true,
            nodes: [
                { id: 'test-action-0', enabled: true, recipients: { nodes: [{ id: 'baz-0', url: '/user/test' }] } },
            ],
        },
        trigger: { id: 'test-0', query: 'test' },
    },
}

export const mockCodeMonitorNodes = [
    {
        id: 'foo0',
        description: 'Test code monitor',
        enabled: true,
        actions: {
            id: 'test-0',
            enabled: true,
            nodes: [{ id: 'test-action-0 ', enabled: true, recipients: { nodes: [{ id: 'baz-0' }] } }],
        },
        trigger: { id: 'test-0', query: 'test' },
    },
    {
        id: 'foo1',
        description: 'Second test code monitor',
        enabled: true,
        actions: {
            id: 'test-1',
            enabled: true,
            nodes: [{ id: 'test-action-1 ', enabled: true, recipients: { nodes: [{ id: 'baz-1' }] } }],
        },
        trigger: { id: 'test-1', query: 'test' },
    },
    {
        id: 'foo2',
        description: 'Third test code monitor',
        enabled: true,
        actions: {
            id: 'test-2',
            enabled: true,
            nodes: [{ id: 'test-action-2 ', enabled: true, recipients: { nodes: [{ id: 'baz-2' }] } }],
        },
        trigger: { id: 'test-2', query: 'test' },
    },
    {
        id: 'foo3',
        description: 'Fourth test code monitor',
        enabled: true,
        actions: {
            id: 'test-3',
            enabled: true,
            nodes: [{ id: 'test-action-3 ', enabled: true, recipients: { nodes: [{ id: 'baz-3' }] } }],
        },
        trigger: { id: 'test-3', query: 'test' },
    },
    {
        id: 'foo4',
        description: 'Fifth test code monitor',
        enabled: true,
        actions: {
            id: 'test-4',
            enabled: true,
            nodes: [{ id: 'test-action-4 ', enabled: true, recipients: { nodes: [{ id: 'baz-4' }] } }],
        },
        trigger: { id: 'test-4', query: 'test' },
    },
    {
        id: 'foo5',
        description: 'Sixth test code monitor',
        enabled: true,
        actions: {
            id: 'test-5',
            enabled: true,
            nodes: [{ id: 'test-action-5 ', enabled: true, recipients: { nodes: [{ id: 'baz-5' }] } }],
        },
        trigger: { id: 'test-5', query: 'test' },
    },
    {
        id: 'foo6',
        description: 'Seventh test code monitor',
        enabled: true,
        actions: {
            id: 'test-6',
            enabled: true,
            nodes: [{ id: 'test-action-6 ', enabled: true, recipients: { nodes: [{ id: 'baz-6' }] } }],
        },
        trigger: { id: 'test-6', query: 'test' },
    },
    {
        id: 'foo7',
        description: 'Eighth test code monitor',
        enabled: true,
        actions: {
            id: 'test-7',
            enabled: true,
            nodes: [{ id: 'test-action-7 ', enabled: true, recipients: { nodes: [{ id: 'baz-7' }] } }],
        },
        trigger: { id: 'test-7', query: 'test' },
    },
    {
        id: 'foo9',
        description: 'Ninth test code monitor',
        enabled: true,
        actions: {
            id: 'test-9',
            enabled: true,
            nodes: [{ id: 'test-action-9 ', enabled: true, recipients: { nodes: [{ id: 'baz-9' }] } }],
        },
        trigger: { id: 'test-9', query: 'test' },
    },
    {
        id: 'foo10',
        description: 'Tenth test code monitor',
        enabled: true,
        actions: {
            id: 'test-0',
            enabled: true,
            nodes: [{ id: 'test-action-0 ', enabled: true, recipients: { nodes: [{ id: 'baz-0' }] } }],
        },
        trigger: { id: 'test-0', query: 'test' },
    },
    {
        id: 'foo11',
        description: 'Eleventh test code monitor',
        enabled: true,
        actions: {
            id: 'test-1',
            enabled: true,
            nodes: [{ id: 'test-action-1 ', enabled: true, recipients: { nodes: [{ id: 'baz-1' }] } }],
        },
        trigger: { id: 'test-1', query: 'test' },
    },
    {
        id: 'foo12',
        description: 'Twelfth test code monitor',
        enabled: true,
        actions: {
            id: 'test-2',
            enabled: true,
            nodes: [{ id: 'test-action-2 ', enabled: true, recipients: { nodes: [{ id: 'baz-2' }] } }],
        },
        trigger: { id: 'test-2', query: 'test' },
    },
]
