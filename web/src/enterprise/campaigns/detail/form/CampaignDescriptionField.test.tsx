import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignDescriptionField } from './CampaignDescriptionField'

describe('CampaignDescriptionField', () => {
    test('renders', () =>
        expect(renderer.create(<CampaignDescriptionField value="a" onChange={() => undefined} />)).toMatchSnapshot())
})
