import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignsMarketing } from './CampaignsMarketing'

describe('CampaignsMarketing', () => {
    test('renders', () => {
        const result = renderer.create(<CampaignsMarketing body={<div>MY CUSTOM CONTENT</div>} />)
        expect(result.toJSON()).toMatchSnapshot()
    })
})
