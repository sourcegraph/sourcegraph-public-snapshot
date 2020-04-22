import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignBranchField } from './CampaignBranchField'

describe('CampaignBranchField', () => {
    test('renders', () =>
        expect(renderer.create(<CampaignBranchField value="a" onChange={() => undefined} />)).toMatchSnapshot())
})
