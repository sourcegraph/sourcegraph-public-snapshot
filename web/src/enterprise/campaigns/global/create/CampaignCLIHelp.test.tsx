import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignCLIHelp } from './CampaignCLIHelp'

describe('CampaignCLIHelp', () => {
    test('renders', () => {
        const result = renderer.create(<CampaignCLIHelp className="test" />)
        expect(result.toJSON()).toMatchSnapshot()
    })
})
