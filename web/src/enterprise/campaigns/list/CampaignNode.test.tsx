import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignNode } from './CampaignNode'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { setLinkComponent } from '../../../../../shared/src/components/Link'

describe('CampaignNode', () => {
    setLinkComponent((props: any) => <a {...props} />)
    afterAll(() => {
        setLinkComponent(null as any)
    })

    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    const node = {
        __typename: 'Campaign',
        id: '123',
        name: 'Upgrade lodash to v4',
        description: `
# Removes lodash

- and renders in markdown
        `,
    } as GQL.ICampaign

    test('renders a campaign node', () => {
        expect(renderer.create(<CampaignNode node={node} />).toJSON()).toMatchSnapshot()
    })
})
