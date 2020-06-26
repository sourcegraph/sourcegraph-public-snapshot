import React from 'react'
import { CampaignsUserMarketingPage } from './CampaignsUserMarketingPage'
import { mount } from 'enzyme'

describe('CampaignsUserMarketingPage', () => {
    test('renders for disabled', () => {
        expect(mount(<CampaignsUserMarketingPage enableReadAccess={false} />).children()).toMatchSnapshot()
    })
    test('renders for enabled', () => {
        expect(mount(<CampaignsUserMarketingPage enableReadAccess={true} />).children()).toMatchSnapshot()
    })
})
