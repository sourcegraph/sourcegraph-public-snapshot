import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignNode } from './CampaignNode'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { parseISO } from 'date-fns'

jest.mock('../icons', () => ({ CampaignsIcon: 'CampaignsIcon' }))

describe('CampaignNode', () => {
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    const node = {
        __typename: 'Campaign',
        id: '123',
        name: 'Upgrade lodash to v4',
        description: `
# Removes lodash

- and renders in markdown
        `,
        changesets: { totalCount: 4, nodes: [{ state: GQL.ChangesetState.OPEN }] },
        changesetPlans: { totalCount: 2 },
        createdAt: '2019-12-04T23:15:01Z',
    } as GQL.ICampaign

    test('renders a campaign node', () => {
        expect(
            renderer.create(<CampaignNode node={node} now={parseISO('2019-01-01T23:15:01Z')} />).toJSON()
        ).toMatchSnapshot()
    })
})
