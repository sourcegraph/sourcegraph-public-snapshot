import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignsUserMarketingPage } from './CampaignsUserMarketingPage'

describe('CampaignsUserMarketingPage', () => {
    test('renders for disabled', () => {
        const result = renderer.create(<CampaignsUserMarketingPage enableReadAccess={false} />)
        expect(result.toJSON()).toMatchSnapshot()
    })
    test('renders for enabled', () => {
        const result = renderer.create(<CampaignsUserMarketingPage enableReadAccess={true} />)
        expect(result.toJSON()).toMatchSnapshot()
    })
})
