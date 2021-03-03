import { subDays } from 'date-fns'
import { ListCampaign } from '../../../graphql-operations'

export const now = new Date()

export const nodes: Record<string, ListCampaign> = {
    'Open campaign': {
        id: 'test',
        url: '/users/alice/campaigns/test',
        name: 'Awesome campaign',
        description: `# What this does

This is my thorough explanation. And it can also get very long, in that case the UI doesn't break though, which is good. And one more line to finally be longer than the viewport.`,
        createdAt: subDays(now, 5).toISOString(),
        closedAt: null,
        changesetsStats: {
            open: 10,
            closed: 0,
            merged: 5,
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
    },
    'No description': {
        id: 'test2',
        url: '/users/alice/campaigns/test2',
        name: 'Awesome campaign',
        description: null,
        createdAt: subDays(now, 5).toISOString(),
        closedAt: null,
        changesetsStats: {
            open: 10,
            closed: 0,
            merged: 5,
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
    },
    'Closed campaign': {
        id: 'test3',
        url: '/users/alice/campaigns/test3',
        name: 'Awesome campaign',
        description: `# My campaign

        This is my thorough explanation.`,
        createdAt: subDays(now, 5).toISOString(),
        closedAt: subDays(now, 3).toISOString(),
        changesetsStats: {
            open: 0,
            closed: 10,
            merged: 5,
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
    },
}
