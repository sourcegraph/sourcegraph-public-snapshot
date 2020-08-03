import React from 'react'
import { CampaignsSiteAdminMarketingPage } from './CampaignsSiteAdminMarketingPage'
import { mount } from 'enzyme'

describe('CampaignsSiteAdminMarketingPage', () => {
    test('renders', () => {
        expect(mount(<CampaignsSiteAdminMarketingPage />)).toMatchSnapshot()
    })
})
