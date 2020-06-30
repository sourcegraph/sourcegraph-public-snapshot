import React from 'react'
import { CreateCampaign } from './CreateCampaign'
import { mount } from 'enzyme'

describe('CreateCampaign', () => {
    test('renders', () => {
        expect(mount(<CreateCampaign />)).toMatchSnapshot()
    })
})
