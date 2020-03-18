import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { GlobalCampaignListPage } from './GlobalCampaignListPage'
import { IUser } from '../../../../../../shared/src/graphql/schema'

describe('GlobalCampaignListPage', () => {
    test('renders for siteadmin', () => {
        const renderer = createRenderer()
        renderer.render(
            <GlobalCampaignListPage
                history={undefined as any}
                location={undefined as any}
                authenticatedUser={{ siteAdmin: true } as IUser}
            />
        )
        expect(renderer.getRenderOutput()).toMatchSnapshot()
    })
    test('renders for non-siteadmin', () => {
        const renderer = createRenderer()
        renderer.render(
            <GlobalCampaignListPage
                history={undefined as any}
                location={undefined as any}
                authenticatedUser={{ siteAdmin: false } as IUser}
            />
        )
        expect(renderer.getRenderOutput()).toMatchSnapshot()
    })
})
