import React from 'react'
import { CampaignBranchField } from './CampaignBranchField'
import { mount } from 'enzyme'

describe('CampaignBranchField', () => {
    test('renders', () =>
        expect(mount(<CampaignBranchField value="a" onChange={() => undefined} />).children()).toMatchSnapshot())
})
