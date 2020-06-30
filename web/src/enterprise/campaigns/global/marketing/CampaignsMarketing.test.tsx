import React from 'react'
import { CampaignsMarketing } from './CampaignsMarketing'
import { mount } from 'enzyme'

describe('CampaignsMarketing', () => {
    test('renders', () => {
        expect(mount(<CampaignsMarketing body={<div>MY CUSTOM CONTENT</div>} />)).toMatchSnapshot()
    })
})
