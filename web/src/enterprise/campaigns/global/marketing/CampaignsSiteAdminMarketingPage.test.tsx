import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignsSiteAdminMarketingPage } from './CampaignsSiteAdminMarketingPage'

describe('CampaignsSiteAdminMarketingPage', () => {
    test('renders', () => {
        const result = renderer.create(<CampaignsSiteAdminMarketingPage />)
        expect(result.toJSON()).toMatchSnapshot()
    })
})
