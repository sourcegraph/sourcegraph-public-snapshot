import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignsDotComPage } from './CampaignsDotComPage'

describe('CampaignsDotComPage', () => {
    test('renders', () => {
        const result = renderer.create(<CampaignsDotComPage />)
        expect(result.toJSON()).toMatchSnapshot()
    })
})
