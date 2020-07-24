import React from 'react'
import { CampaignNode } from './CampaignNode'
import { parseISO } from 'date-fns'
import { createMemoryHistory } from 'history'
import { mount } from 'enzyme'

const now = parseISO('2019-01-01T23:15:01Z')

jest.mock('../icons', () => ({ CampaignsIcon: 'CampaignsIcon' }))

describe('CampaignNode', () => {
    const node = {
        __typename: 'Campaign',
        id: '123',
        name: 'Upgrade lodash to v4',
        description: `
# Removes lodash

- and renders in markdown
        `,
        changesets: { stats: { merged: 0, open: 1, closed: 3 } },
        patches: { totalCount: 2 },
        createdAt: '2019-12-04T23:15:01Z',
        closedAt: null,
        author: {
            username: 'alice',
        },
    }

    test('open campaign', () => {
        expect(mount(<CampaignNode node={node} now={now} history={createMemoryHistory()} />)).toMatchSnapshot()
    })
    test('closed campaign', () => {
        expect(
            mount(
                <CampaignNode
                    node={{ ...node, closedAt: '2019-12-04T23:19:01Z' }}
                    now={now}
                    history={createMemoryHistory()}
                />
            )
        ).toMatchSnapshot()
    })
    test('campaign without description', () => {
        expect(
            mount(
                <CampaignNode
                    node={{
                        ...node,
                        description: null,
                    }}
                    now={now}
                    history={createMemoryHistory()}
                />
            )
        ).toMatchSnapshot()
    })
})
