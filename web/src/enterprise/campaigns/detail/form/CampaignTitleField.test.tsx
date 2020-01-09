import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignTitleField } from './CampaignTitleField'

describe('CampaignTitleField', () => {
    test('renders', () =>
        expect(renderer.create(<CampaignTitleField value="a" onChange={() => undefined} />)).toMatchSnapshot())
})
