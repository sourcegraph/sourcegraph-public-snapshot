import React from 'react'
import { CampaignTitleField } from './CampaignTitleField'
import { mount } from 'enzyme'

describe('CampaignTitleField', () => {
    test('renders', () => expect(mount(<CampaignTitleField value="a" onChange={() => undefined} />)).toMatchSnapshot())
})
