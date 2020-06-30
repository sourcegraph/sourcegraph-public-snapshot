import React from 'react'
import { CampaignDescriptionField } from './CampaignDescriptionField'
import { mount } from 'enzyme'

describe('CampaignDescriptionField', () => {
    test('renders', () =>
        expect(mount(<CampaignDescriptionField value="a" onChange={() => undefined} />)).toMatchSnapshot())
})
